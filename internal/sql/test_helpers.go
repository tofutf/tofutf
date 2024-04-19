package sql

import (
	"context"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/sql/postgres"
	"github.com/tofutf/tofutf/internal/xslog"

	"github.com/jackc/pgx/v5"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

// TestingT is used instead of directly relying on testing.T to prevent the
// package from being bundled in a normal release build of tofutf.
type TestingT interface {
	require.TestingT
	assert.TestingT
	Helper()
	Name() string
	Skip(...any)
	Cleanup(f func())
}

// NewTestDB creates a logical database in postgres for a test and returns a
// connection string for connecting to the database. The database is dropped
// upon test completion.
func NewTestDB(t TestingT) string {
	t.Helper()

	connstr, ok := os.LookupEnv(TestDatabaseURL)
	if !ok {
		t.Skip("Export valid OTF_TEST_DATABASE_URL before running this test")
	}

	ctx := context.Background()

	// connect and create database
	conn, err := pgx.Connect(ctx, connstr)
	require.NoError(t, err)

	// generate a safe, unique logical database name
	logical := t.Name()
	logical = strcase.ToSnake(logical)
	logical = strings.ReplaceAll(logical, "/", "_")
	// NOTE: maximum size of a postgres name is 31
	// 21 + "_" + 8 = 30
	if len(logical) > 22 {
		logical = logical[:22]
	}
	logical = logical + "_" + strings.ToLower(internal.GenerateRandomString(8))

	_, err = conn.Exec(ctx, "CREATE DATABASE "+logical)
	require.NoError(t, err, "unable to create database")
	t.Cleanup(func() {
		_, err := conn.Exec(ctx, "DROP DATABASE "+logical)
		assert.NoError(t, err, "unable to drop database %s", logical)
		err = conn.Close(ctx)
		assert.NoError(t, err, "unable to close connection")
	})

	// modify connection string to use new logical database
	u, err := url.Parse(connstr)
	require.NoError(t, err)
	u.Path = "/" + logical

	return u.String()
}

// NewTestContainer returns a new test container.
func NewTestContainer(t TestingT) *postgres.PostgresContainer {
	t.Helper()

	ctx := context.Background()
	postgresContainer, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("tofutf"),
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

	logger := slog.New(&xslog.NoopHandler{})

	pool, err := New(ctx, Options{
		Logger:     logger,
		ConnString: connStr,
	})
	require.NoError(t, err)

	err = postgresContainer.Snapshot(ctx, postgres.WithSnapshotName("root"))
	require.NoError(t, err)

	pool.Close()

	return postgresContainer
}

// NewTestContainerPool returns a new Pool that is connected to the returned PostgresContainer.
func NewTestContainerPool(t TestingT) (*postgres.PostgresContainer, *Pool) {
	t.Helper()

	pg := NewTestContainer(t)

	connStr, err := pg.ConnectionString(context.Background())
	require.NoError(t, err)

	pool, err := New(context.Background(), Options{
		Logger:     slog.New(&xslog.NoopHandler{}),
		ConnString: connStr,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	return pg, pool
}

// TestContainerReset resets the test container and provides a new pool that connects to it.
func TestContainerReset(t TestingT, pg *postgres.PostgresContainer) *Pool {
	t.Helper()

	ctx := context.Background()

	err := pg.Restore(ctx, postgres.WithSnapshotName("root"))
	require.NoError(t, err)

	connStr, err := pg.ConnectionString(context.Background())
	require.NoError(t, err)

	pool, err := New(context.Background(), Options{
		Logger:     slog.Default(),
		ConnString: connStr,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}
