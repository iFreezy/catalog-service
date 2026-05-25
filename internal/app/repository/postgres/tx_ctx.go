package rcpostgres

import (
	"context"

	"github.com/uptrace/bun"
)

type _ctxKeyTx struct{}

func getTxFromContext(ctx context.Context) bun.Tx {
	tx, _ := ctx.Value(_ctxKeyTx{}).(bun.Tx)
	return tx
}

func setTxToContext(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, _ctxKeyTx{}, tx)
}
