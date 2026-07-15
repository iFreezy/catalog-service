package repository

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/iFreezy/catalog-service/internal/app/entity"
)

type (
	Transactional interface {
		InsideTx(ctx context.Context, f func(ctx context.Context) error) error
	}

	Migrate interface {
		Migrate(ctx context.Context) (oldVer, newVer int64, err error)
	}

	Category interface {
		Transactional
		Create(ctx context.Context, category entity.Category) error
		GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Category, error)
		Update(ctx context.Context, category entity.Category) error
		Delete(ctx context.Context, guid uuid.UUID) error
		List(ctx context.Context, name *string) ([]entity.Category, error)
	}

	Product interface {
		Transactional
		Create(ctx context.Context, product entity.Product) error
		GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Product, error)
		Update(ctx context.Context, product entity.Product) error
		Delete(ctx context.Context, guid uuid.UUID) error
		List(ctx context.Context, name *string, categoryGUID *uuid.UUID) ([]entity.Product, error)
	}
)
