package ghcatalogv1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/iFreezy/catalog-service/internal/app/entity"
	"github.com/iFreezy/catalog-service/internal/app/service"
	catalogv1 "github.com/iFreezy/catalog-service/internal/pkg/grpc/gen/catalog/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type productServiceStub struct {
	service.Product
	products []entity.Product
	err      error
}

func (s *productServiceStub) GetByGUIDs(_ context.Context, _ []uuid.UUID) ([]entity.Product, error) {
	return s.products, s.err
}

func TestGetProduct(t *testing.T) {
	guid := uuid.Must(uuid.NewV4())
	product := entity.Product{GUID: guid, Name: "Book", Price: 12.5, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	tests := []struct {
		name string
		guid string
		svc  *productServiceStub
		code codes.Code
	}{
		{"invalid", "bad", &productServiceStub{}, codes.InvalidArgument},
		{"missing", guid.String(), &productServiceStub{}, codes.NotFound},
		{"internal", guid.String(), &productServiceStub{err: errors.New("db unavailable")}, codes.Internal},
		{"found", guid.String(), &productServiceStub{products: []entity.Product{product}}, codes.OK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := NewHandler(tc.svc).GetProduct(context.Background(), &catalogv1.GetProductRequest{Guid: tc.guid})
			require.Equal(t, tc.code, status.Code(err))
			if tc.code == codes.OK {
				require.Equal(t, guid.String(), resp.GetProduct().GetGuid())
				require.EqualValues(t, 1250, resp.GetProduct().GetPrice())
			}
		})
	}
}

func TestGetProductsReportsMissing(t *testing.T) {
	found := uuid.Must(uuid.NewV4())
	missing := uuid.Must(uuid.NewV4())
	svc := &productServiceStub{products: []entity.Product{{GUID: found}}}
	resp, err := NewHandler(svc).GetProducts(context.Background(), &catalogv1.GetProductsRequest{Guids: []string{found.String(), missing.String()}})
	require.NoError(t, err)
	require.Len(t, resp.GetProducts(), 1)
	require.Equal(t, []string{missing.String()}, resp.GetMissingGuids())
}

func TestGetProductsRejectsInvalidGUID(t *testing.T) {
	_, err := NewHandler(&productServiceStub{}).GetProducts(context.Background(), &catalogv1.GetProductsRequest{Guids: []string{"bad"}})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
