package pproduct

import (
	"context"
	"database/sql"

	"github.com/gofrs/uuid"
	"github.com/iFreezy/catalog-service/internal/app/entity"
	"github.com/iFreezy/catalog-service/internal/app/repository"
	rcconn "github.com/iFreezy/catalog-service/internal/app/repository/conn/postgres"
	rcpostgres "github.com/iFreezy/catalog-service/internal/app/repository/postgres"
	"github.com/iFreezy/catalog-service/internal/app/util"
)

type repo struct {
	*rcpostgres.Client
}

var _ repository.Product = (*repo)(nil)

func NewRepoFromPostgres(client *rcpostgres.Client) repository.Product {
	return &repo{Client: client}
}

func (r *repo) Create(ctx context.Context, product entity.Product) error {
	_, err := r.NewInsert().Model(&product).Exec(ctx)
	return err
}

func (r *repo) GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Product, error) {
	var product entity.Product
	err := r.NewSelect().Model(&product).Where("guid = ?", guid).Scan(ctx)
	if err != nil {
		return entity.Product{}, util.ReplaceErr1(err, sql.ErrNoRows, entity.ErrNotFound)
	}
	return product, nil
}

func (r *repo) Update(ctx context.Context, product entity.Product) error {
	res, err := r.NewUpdate().Model(&product).WherePK().Exec(ctx)
	return rcconn.UpdateErr(res, err)
}

func (r *repo) Delete(ctx context.Context, guid uuid.UUID) error {
	_, err := r.NewDelete().Model((*entity.Product)(nil)).Where("guid = ?", guid).Exec(ctx)
	return rcconn.DeleteErr(err)
}

func (r *repo) List(ctx context.Context, name *string, categoryGUID *uuid.UUID) ([]entity.Product, error) {
	var products []entity.Product
	q := r.NewSelect().Model(&products)
	if name != nil {
		q = q.Where("name = ?", *name)
	}
	if categoryGUID != nil {
		q = q.Where("category_guid = ?", *categoryGUID)
	}
	err := q.Scan(ctx)
	return products, rcconn.DeleteErr(err)
}
