package pcategory

import (
	"context"
	"database/sql"

	"github.com/gofrs/uuid"
	"github.com/iFreezy/catalog-service/internal/app/entity"
	"github.com/iFreezy/catalog-service/internal/app/repository"
	rcpostgres "github.com/iFreezy/catalog-service/internal/app/repository/conn/postgres"
	"github.com/iFreezy/catalog-service/internal/app/util"
	"github.com/uptrace/bun"
)

type repo struct {
	db bun.IDB
}

func NewRepoFromPostgres(db bun.IDB) repository.Category {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, category entity.Category) error {
	_, err := r.db.NewInsert().Model(&category).Exec(ctx)
	return err
}

func (r *repo) GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Category, error) {
	var category entity.Category
	err := r.db.NewSelect().Model(&category).Where("guid = ?", guid).Scan(ctx)
	if err != nil {
		return entity.Category{}, util.ReplaceErr1(err, sql.ErrNoRows, entity.ErrNotFound)
	}
	return category, nil
}

func (r *repo) Update(ctx context.Context, category entity.Category) error {
	res, err := r.db.NewUpdate().Model(&category).WherePK().Exec(ctx)
	return rcpostgres.UpdateErr(res, err)
}

func (r *repo) Delete(ctx context.Context, guid uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*entity.Category)(nil)).Where("guid = ?", guid).Exec(ctx)
	return rcpostgres.DeleteErr(err)
}

func (r *repo) List(ctx context.Context, name *string) ([]entity.Category, error) {
	var categories []entity.Category
	q := r.db.NewSelect().Model(&categories)
	if name != nil {
		q = q.Where("name = ?", *name)
	}
	err := q.Scan(ctx)
	return categories, err
}
