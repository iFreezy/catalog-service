package pcategory

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

var _ repository.Category = (*repo)(nil)

func NewRepoFromPostgres(client *rcpostgres.Client) repository.Category {
	return &repo{Client: client}
}

func (r *repo) Create(ctx context.Context, category entity.Category) error {
	_, err := r.NewInsert().Model(&category).Exec(ctx)
	return err
}

func (r *repo) GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Category, error) {
	var category entity.Category
	err := r.NewSelect().Model(&category).Where("guid = ?", guid).Scan(ctx)
	if err != nil {
		return entity.Category{}, util.ReplaceErr1(err, sql.ErrNoRows, entity.ErrNotFound)
	}
	return category, nil
}

func (r *repo) Update(ctx context.Context, category entity.Category) error {
	res, err := r.NewUpdate().Model(&category).WherePK().Exec(ctx)
	return rcconn.UpdateErr(res, err)
}

func (r *repo) Delete(ctx context.Context, guid uuid.UUID) error {
	_, err := r.NewDelete().Model((*entity.Category)(nil)).Where("guid = ?", guid).Exec(ctx)
	return rcconn.DeleteErr(err)
}

func (r *repo) List(ctx context.Context, name *string) ([]entity.Category, error) {
	var categories []entity.Category
	q := r.NewSelect().Model(&categories)
	if name != nil {
		q = q.Where("name = ?", *name)
	}
	err := q.Scan(ctx)
	return categories, err
}
