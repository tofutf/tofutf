package sql_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/sql"
)

var samplePgErr1 = &pgconn.PgError{
	Code: "23503",
}
var samplePgErr2 = &pgconn.PgError{
	Code: "23505",
}

var samplePgErr3 = &pgconn.PgError{
	Code: "23599",
}

func TestSqlErrorHandling(t *testing.T) {

	tests := []struct {
		name string
		err  error
		want error
	}{
		{"Expect no rows in result set without wrapping", errors.New("no rows in result set"), internal.ErrResourceNotFound},
		{"Expect no rows in result set with wrapping", fmt.Errorf("something else: %w", errors.New("no rows in result set")), internal.ErrResourceNotFound},
		{"Expect handling of PG 23503", samplePgErr1, &internal.ForeignKeyError{PgError: samplePgErr1}},
		{"Expect handling of PG 23505", samplePgErr2, internal.ErrResourceAlreadyExists},
		{"Expect raw return of other PG codes", samplePgErr3, samplePgErr3},
		{"Expect raw return for any other error", internal.ErrAccessNotPermitted, internal.ErrAccessNotPermitted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sql.Error(tt.err)
			assert.Equal(t, tt.want, err)
		})
	}
}
