package sql

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

func TestSetDefaultMaxConnections(t *testing.T) {
	tests := []struct {
		name    string
		connstr string
		want    string
	}{
		{"empty dsn", "", "pool_max_conns=20"},
		{"non-empty dsn", "user=louis host=localhost", "user=louis host=localhost pool_max_conns=20"},
		{"postgres url", "postgres:///otf", "postgres:///otf?pool_max_conns=20"},
		{"postgresql url", "postgresql:///otf", "postgresql:///otf?pool_max_conns=20"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setDefaultMaxConnections(tt.connstr, 20)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPool(t *testing.T) {
	ctx := context.Background()
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	// Clean up the container
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	})

	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// TODO(johnrowl): use snapshots here to reset state
	// when: https://github.com/testcontainers/testcontainers-go/issues/2474 is fixed
	t.Run("Tx", func(t *testing.T) {
		t.Run("should commit changes when think passes", func(t *testing.T) {
			pool, err := New(ctx, Options{
				Logger:     slog.Default(),
				ConnString: connStr,
			})
			require.NoError(t, err)

			defer pool.Close()

			err = pool.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.InsertAgent(ctx, pggen.InsertAgentParams{
					AgentID:      String("id1"),
					Name:         String("name1"),
					Version:      String("v1.0.0"),
					MaxJobs:      Int4(3),
					IPAddress:    net.IPv4(192, 168, 1, 100),
					LastPingAt:   Timestamptz(time.Now()),
					LastStatusAt: Timestamptz(time.Now()),
					Status:       String("idle"),
					AgentPoolID:  NullString(),
				})
				return err
			})
			require.NoError(t, err)

			// the previous transaction should have been committed.
			err = pool.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.FindAgentByID(ctx, String("id"))
				return err
			})
			require.NoError(t, err)
		})

		t.Run("should rollback changes when thunk fails", func(t *testing.T) {
			pool, err := New(ctx, Options{
				Logger:     slog.Default(),
				ConnString: connStr,
			})
			require.NoError(t, err)

			defer pool.Close()

			err = pool.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.InsertAgent(ctx, pggen.InsertAgentParams{
					AgentID:      String("id2"),
					Name:         String("name2"),
					Version:      String("v1.0.0"),
					MaxJobs:      Int4(3),
					IPAddress:    net.IPv4(192, 168, 1, 100),
					LastPingAt:   Timestamptz(time.Now()),
					LastStatusAt: Timestamptz(time.Now()),
					Status:       String("idle"),
					AgentPoolID:  NullString(),
				})
				require.NoError(t, err)

				_, err = q.FindAgentByID(ctx, String("id"))
				require.NoError(t, err)

				return fmt.Errorf("fake error")
			})
			require.Error(t, err)

			// the previous transaction should have been rolled back.
			err = pool.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.FindAgentByID(ctx, String("id"))
				return err
			})
			require.Error(t, err)
		})
	})
}
