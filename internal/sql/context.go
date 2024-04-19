package sql

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const (
	// context key for retrieving connection from context
	connCtxKey ctxKey = 1
	txCtxKey   ctxKey = 2
)

type ctxKey int

func newContext(ctx context.Context, conn *pgx.Conn) context.Context {
	return context.WithValue(ctx, connCtxKey, conn)
}

func fromContext(ctx context.Context) (*pgx.Conn, bool) {
	conn, ok := ctx.Value(connCtxKey).(*pgx.Conn)
	return conn, ok
}

func newTxContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey, tx)
}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey).(pgx.Tx)
	return tx, ok
}
