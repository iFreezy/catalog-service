package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/iFreezy/catalog-service/internal/app/config/section"
	"github.com/iFreezy/catalog-service/migration"
	"github.com/rs/zerolog/log"
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

	log.Debug().
		Stringer("read_timeout", cfg.ReadTimeout).
		Stringer("write_timeout", cfg.WriteTimeout).
		Msg("PostgreSQL connection params")

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

	bunDB := newBunIdbTxInjector(rawBunDB)

	return &Client{
		_bunDB:   bunDB,
		rawBunDB: rawBunDB,
		cfg:      cfg,
	}, nil
}

func (c *Client) InsideTx(ctx context.Context, f func(ctx context.Context) error) error {
	if tx := getTxFromContext(ctx); tx.Tx != nil {
		log.Ctx(ctx).Debug().Msg("already in transaction, reusing")
		return f(ctx)
	}

	tx, err := c.rawBunDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	done := false
	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()

	log.Ctx(ctx).Debug().Msg("starting transaction")

	ctx = setTxToContext(ctx, tx)

	if err = f(ctx); err != nil {
		return err
	}

	done = true
	log.Ctx(ctx).Debug().Msg("committing transaction")

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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
