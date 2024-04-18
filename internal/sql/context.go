package sql

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const (
	// context key for retrieving connection from context
	connCtxKey ctxKey = 1
)

type ctxKey int

func newContext(ctx context.Context, conn *pgx.Conn) context.Context {
	return context.WithValue(ctx, connCtxKey, conn)
}

func fromContext(ctx context.Context) (*pgx.Conn, bool) {
	conn, ok := ctx.Value(connCtxKey).(*pgx.Conn)
	return conn, ok
}
