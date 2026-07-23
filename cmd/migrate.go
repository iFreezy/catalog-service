package cmd

import (
	"strings"

	"github.com/iFreezy/catalog-service/internal/app/builder"
	"github.com/urfave/cli/v2"
)

const (
	cmdMigrateUsage       = "Применяет миграции базы данных при наличии новых"
	cmdMigrateDescription = `
Устанавливает соединение к Postgres базе данных, проверяет соединение,
и затем применяет те миграции, которые еще не были применены,
в соответствии со схемой данных.
`
)

func Migrate() *cli.Command {
	return &cli.Command{
		Name:            "migrate",
		Aliases:         []string{"m"},
		Usage:           cmdMigrateUsage,
		Description:     strings.TrimSpace(cmdMigrateDescription),
		Action:          cmdMigrate,
		HideHelpCommand: true,
	}
}

func cmdMigrate(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)

	app.BuildConfig()
	app.BuildRepoConnPostgres()
	app.BuildRepoConnMigrator()

	app.Run()

	return nil
}
