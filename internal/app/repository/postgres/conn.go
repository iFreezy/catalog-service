package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/iFreezy/catalog-service/internal/app/config/section"
	"github.com/iFreezy/catalog-service/migration"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

type (
	Client struct {
		_bunDB
		rawBunDB *bun.DB

		cfg section.RepositoryPostgres
	}

	_bunDB = bun.IDB
)

func (c *Client) GetRawBunDB() *bun.DB {
	return c.rawBunDB
}

func NewConn(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	var u url.URL
	u.Scheme = "postgres"
	u.Host = cfg.Address
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	u.Path = cfg.Name

	var args = make(url.Values)
	args.Set("sslmode", "disable")
	u.RawQuery = args.Encode()

	log.Printf("PostgreSQL connection: ReadTimeout=%s, WriteTimeout=%s", cfg.ReadTimeout, cfg.WriteTimeout)

	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(u.String()),
		pgdriver.WithReadTimeout(cfg.ReadTimeout),
		pgdriver.WithWriteTimeout(cfg.WriteTimeout),
	))

	sqlDB.SetMaxOpenConns(10)

	rawBunDB := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	var cancelFunc func()
	ctx, cancelFunc = context.WithTimeout(ctx, 2*time.Second)
	defer cancelFunc()

	if err := rawBunDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping connection: %w", err)
	}

	return &Client{
		_bunDB:   rawBunDB,
		rawBunDB: rawBunDB,
		cfg:      cfg,
	}, nil
}

func (c *Client) Migrate(ctx context.Context) (oldVer, newVer int64, err error) {
	migrations := migrate.NewMigrations()

	if err = migrations.Discover(migration.Postgres); err != nil {
		return 0, 0, fmt.Errorf("failed to discover migrations: %w", err)
	}

	opts := []migrate.MigratorOption{
		migrate.WithTableName(c.cfg.MigrationTable),
		migrate.WithLocksTableName(c.cfg.MigrationTable + "_lock"),
		migrate.WithMarkAppliedOnSuccess(true),
	}

	m := migrate.NewMigrator(c.rawBunDB, migrations, opts...)

	if err = m.Init(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to init migrations table: %w", err)
	}

	applied, err := m.AppliedMigrations(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) > 0 {
		oldVer, _ = strconv.ParseInt(applied[0].Name, 10, 64)
	}

	mgg, err := m.Migrate(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to apply migrations: %w", err)
	}

	newVer = oldVer
	for _, mg := range mgg.Migrations {
		ver, _ := strconv.ParseInt(mg.Name, 10, 64)
		if ver > newVer {
			newVer = ver
		}
	}

	return oldVer, newVer, nil
}
