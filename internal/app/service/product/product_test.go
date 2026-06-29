package sproduct_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iFreezy/catalog-service/internal/app/entity"
	"github.com/iFreezy/catalog-service/internal/app/repository/mocks"
	"github.com/iFreezy/catalog-service/internal/app/service"
	sproduct "github.com/iFreezy/catalog-service/internal/app/service/product"
)

var errRepo = errors.New("repo failure")

func ptr(s string) *string { return &s }

type deps struct {
	svc          service.Product
	repoProduct  *mocks.MockProduct
	repoCategory *mocks.MockCategory
}

func newDeps(t *testing.T) deps {
	t.Helper()

	repoProduct := mocks.NewMockProduct(t)
	repoCategory := mocks.NewMockCategory(t)

	return deps{
		svc:          sproduct.NewService(repoProduct, repoCategory),
		repoProduct:  repoProduct,
		repoCategory: repoCategory,
	}
}

func sampleCreateReq() entity.RequestProductCreate {
	return entity.RequestProductCreate{
		Name:         "Coffee",
		Description:  ptr("Fresh beans"),
		Price:        9.99,
		CategoryGUID: uuid.Must(uuid.NewV4()),
	}
}

func sampleUpdateReq() entity.RequestProductUpdate {
	return entity.RequestProductUpdate{
		Name:         "Coffee Updated",
		Description:  ptr("Roasted"),
		Price:        12.5,
		CategoryGUID: uuid.Must(uuid.NewV4()),
	}
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(d deps, req entity.RequestProductCreate)
		wantErr error
	}{
		{
			name: "list lookup fails",
			setup: func(d deps, _ entity.RequestProductCreate) {
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errRepo).Once()
			},
			wantErr: errRepo,
		},
		{
			name: "duplicate name rejected",
			setup: func(d deps, _ entity.RequestProductCreate) {
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return([]entity.Product{{Name: "Coffee"}}, nil).Once()
			},
			wantErr: entity.ErrProductDuplicate,
		},
		{
			name: "category not found",
			setup: func(d deps, req entity.RequestProductCreate) {
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				d.repoCategory.EXPECT().
					GetByGUID(mock.Anything, req.CategoryGUID).
					Return(entity.Category{}, entity.ErrNotFound).Once()
			},
			wantErr: entity.ErrNotFound,
		},
		{
			name: "repository create fails",
			setup: func(d deps, req entity.RequestProductCreate) {
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				d.repoCategory.EXPECT().
					GetByGUID(mock.Anything, req.CategoryGUID).
					Return(entity.Category{GUID: req.CategoryGUID}, nil).Once()
				d.repoProduct.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(errRepo).Once()
			},
			wantErr: errRepo,
		},
		{
			name: "success",
			setup: func(d deps, req entity.RequestProductCreate) {
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				d.repoCategory.EXPECT().
					GetByGUID(mock.Anything, req.CategoryGUID).
					Return(entity.Category{GUID: req.CategoryGUID}, nil).Once()
				d.repoProduct.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(p entity.Product) bool {
						return p.Name == req.Name && p.Price == req.Price &&
							p.CategoryGUID == req.CategoryGUID && p.GUID != uuid.Nil
					})).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := newDeps(t)
			req := sampleCreateReq()
			tc.setup(d, req)

			got, err := d.svc.Create(context.Background(), req)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, entity.Product{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, req.Name, got.Name)
			assert.Equal(t, req.Price, got.Price)
			assert.Equal(t, req.CategoryGUID, got.CategoryGUID)
			assert.NotEqual(t, uuid.Nil, got.GUID)
			assert.False(t, got.CreatedAt.IsZero())
		})
	}
}

func TestService_GetByGUID(t *testing.T) {
	t.Parallel()

	guid := uuid.Must(uuid.NewV4())
	want := entity.Product{GUID: guid, Name: "Coffee"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		d := newDeps(t)
		d.repoProduct.EXPECT().GetByGUID(mock.Anything, guid).Return(want, nil).Once()

		got, err := d.svc.GetByGUID(context.Background(), guid)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		d := newDeps(t)
		d.repoProduct.EXPECT().
			GetByGUID(mock.Anything, guid).
			Return(entity.Product{}, entity.ErrNotFound).Once()

		got, err := d.svc.GetByGUID(context.Background(), guid)

		require.ErrorIs(t, err, entity.ErrNotFound)
		assert.Equal(t, entity.Product{}, got)
	})
}

func TestService_Update(t *testing.T) {
	t.Parallel()

	guid := uuid.Must(uuid.NewV4())

	tests := []struct {
		name    string
		setup   func(d deps, req entity.RequestProductUpdate)
		wantErr error
	}{
		{
			name: "product not found",
			setup: func(d deps, _ entity.RequestProductUpdate) {
				d.repoProduct.EXPECT().
					GetByGUID(mock.Anything, guid).
					Return(entity.Product{}, entity.ErrNotFound).Once()
			},
			wantErr: entity.ErrNotFound,
		},
		{
			name: "list lookup fails",
			setup: func(d deps, _ entity.RequestProductUpdate) {
				d.repoProduct.EXPECT().
					GetByGUID(mock.Anything, guid).
					Return(entity.Product{GUID: guid}, nil).Once()
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errRepo).Once()
			},
			wantErr: errRepo,
		},
		{
			name: "duplicate name on another product",
			setup: func(d deps, _ entity.RequestProductUpdate) {
				d.repoProduct.EXPECT().
					GetByGUID(mock.Anything, guid).
					Return(entity.Product{GUID: guid}, nil).Once()
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return([]entity.Product{{GUID: uuid.Must(uuid.NewV4())}}, nil).Once()
			},
			wantErr: entity.ErrProductDuplicate,
		},
		{
			name: "category not found",
			setup: func(d deps, req entity.RequestProductUpdate) {
				d.repoProduct.EXPECT().
					GetByGUID(mock.Anything, guid).
					Return(entity.Product{GUID: guid}, nil).Once()
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return([]entity.Product{{GUID: guid}}, nil).Once()
				d.repoCategory.EXPECT().
					GetByGUID(mock.Anything, req.CategoryGUID).
					Return(entity.Category{}, entity.ErrNotFound).Once()
			},
			wantErr: entity.ErrNotFound,
		},
		{
			name: "repository update fails",
			setup: func(d deps, req entity.RequestProductUpdate) {
				d.repoProduct.EXPECT().
					GetByGUID(mock.Anything, guid).
					Return(entity.Product{GUID: guid}, nil).Once()
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				d.repoCategory.EXPECT().
					GetByGUID(mock.Anything, req.CategoryGUID).
					Return(entity.Category{GUID: req.CategoryGUID}, nil).Once()
				d.repoProduct.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(errRepo).Once()
			},
			wantErr: errRepo,
		},
		{
			name: "success",
			setup: func(d deps, req entity.RequestProductUpdate) {
				d.repoProduct.EXPECT().
					GetByGUID(mock.Anything, guid).
					Return(entity.Product{GUID: guid, Name: "Old"}, nil).Once()
				d.repoProduct.EXPECT().
					List(mock.Anything, mock.Anything, mock.Anything).
					Return([]entity.Product{{GUID: guid}}, nil).Once()
				d.repoCategory.EXPECT().
					GetByGUID(mock.Anything, req.CategoryGUID).
					Return(entity.Category{GUID: req.CategoryGUID}, nil).Once()
				d.repoProduct.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p entity.Product) bool {
						return p.GUID == guid && p.Name == req.Name &&
							p.CategoryGUID == req.CategoryGUID
					})).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := newDeps(t)
			req := sampleUpdateReq()
			tc.setup(d, req)

			got, err := d.svc.Update(context.Background(), guid, req)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, entity.Product{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, guid, got.GUID)
			assert.Equal(t, req.Name, got.Name)
			assert.Equal(t, req.Price, got.Price)
			assert.Equal(t, req.CategoryGUID, got.CategoryGUID)
			assert.False(t, got.UpdatedAt.IsZero())
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	guid := uuid.Must(uuid.NewV4())

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		d := newDeps(t)
		d.repoProduct.EXPECT().Delete(mock.Anything, guid).Return(nil).Once()

		require.NoError(t, d.svc.Delete(context.Background(), guid))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		d := newDeps(t)
		d.repoProduct.EXPECT().Delete(mock.Anything, guid).Return(errRepo).Once()

		require.ErrorIs(t, d.svc.Delete(context.Background(), guid), errRepo)
	})
}

func TestService_List(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		d := newDeps(t)
		want := []entity.Product{{Name: "A"}, {Name: "B"}}
		d.repoProduct.EXPECT().
			List(mock.Anything, mock.Anything, mock.Anything).
			Return(want, nil).Once()

		got, err := d.svc.List(context.Background())

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		d := newDeps(t)
		d.repoProduct.EXPECT().
			List(mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errRepo).Once()

		got, err := d.svc.List(context.Background())

		require.ErrorIs(t, err, errRepo)
		assert.Nil(t, got)
	})
}
