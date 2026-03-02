package scategory

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/iFreezy/catalog-service/internal/app/entity"
	"github.com/iFreezy/catalog-service/internal/app/repository"
	"github.com/iFreezy/catalog-service/internal/app/service"
)

type svc struct {
	repoCategory repository.Category
	repoProduct  repository.Product
}

func NewService(repoCategory repository.Category, repoProduct repository.Product) service.Category {
	return &svc{
		repoCategory: repoCategory,
		repoProduct:  repoProduct,
	}
}

func (s *svc) Create(ctx context.Context, req entity.RequestCategoryCreate) (entity.Category, error) {
	existing, err := s.repoCategory.List(ctx, &req.Name)
	if err != nil {
		return entity.Category{}, err
	}
	if len(existing) > 0 {
		return entity.Category{}, entity.ErrAlreadyExists
	}

	now := time.Now()
	category := entity.Category{
		GUID:      uuid.Must(uuid.NewV4()),
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repoCategory.Create(ctx, category); err != nil {
		return entity.Category{}, err
	}

	return category, nil
}

func (s *svc) GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Category, error) {
	return s.repoCategory.GetByGUID(ctx, guid)
}

func (s *svc) Update(ctx context.Context, guid uuid.UUID, req entity.RequestCategoryUpdate) (entity.Category, error) {
	category, err := s.repoCategory.GetByGUID(ctx, guid)
	if err != nil {
		return entity.Category{}, err
	}

	existing, err := s.repoCategory.List(ctx, &req.Name)
	if err != nil {
		return entity.Category{}, err
	}
	if len(existing) > 0 && existing[0].GUID != guid {
		return entity.Category{}, entity.ErrAlreadyExists
	}

	category.Name = req.Name
	category.UpdatedAt = time.Now()

	if err := s.repoCategory.Update(ctx, category); err != nil {
		return entity.Category{}, err
	}

	return category, nil
}

func (s *svc) Delete(ctx context.Context, guid uuid.UUID) error {
	if _, err := s.repoCategory.GetByGUID(ctx, guid); err != nil {
		return err
	}

	products, err := s.repoProduct.List(ctx, nil, &guid)
	if err != nil {
		return err
	}
	if len(products) > 0 {
		return entity.ErrCategoryHasProducts
	}

	return s.repoCategory.Delete(ctx, guid)
}

func (s *svc) List(ctx context.Context) ([]entity.Category, error) {
	return s.repoCategory.List(ctx, nil)
}
