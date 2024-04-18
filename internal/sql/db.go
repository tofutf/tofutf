package sql

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

const (
	// max conns avail in a pgx pool
	defaultMaxConnections = 10
)

type (
	// Pool provides access to the postgres db as well as queries generated from
	// SQL
	Pool struct {
		*pgxpool.Pool // db connection pool
		logger        *slog.Logger
	}

	// Options for constructing a DB
	Options struct {
		Logger     *slog.Logger
		ConnString string
	}
)

// New constructs a new DB connection pool, and migrates the schema to the
// latest version.
func New(ctx context.Context, opts Options) (*Pool, error) {
	// Bump max number of connections in a pool. By default pgx sets it to the
	// greater of 4 or the num of CPUs. However, otfd acquires several dedicated
	// connections for session-level advisory locks and can easily exhaust this.
	connString, err := setDefaultMaxConnections(opts.ConnString, defaultMaxConnections)
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	opts.Logger.Info(
		"connected to database",
		"database", config.ConnConfig.Database,
		"host", config.ConnConfig.Host,
		"port", config.ConnConfig.Port,
		"user", config.ConnConfig.User,
	)

	// goose gets upset with max_pool_conns parameter so pass it the unaltered
	// connection string
	if err := migrate(opts.Logger, opts.ConnString); err != nil {
		return nil, err
	}

	return &Pool{
		Pool:   pool,
		logger: opts.Logger,
	}, nil
}

// Query obtains a connection for the pool, executes the given function, and
// returns the connection to the pool.
func Query[T any](ctx context.Context, pool *Pool, fn func(context.Context, pggen.Querier) (T, error)) (T, error) {
	var result T
	err := pool.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		v, err := fn(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to invoke func: %w", err)
		}

		result = v

		return nil
	})
	if err != nil {
		return result, fmt.Errorf("failed to invoke func: %w", err)
	}

	return result, nil
}

// Tx obtains a transaction from the pool, executes the given fn, and then commits the transaction.
func Tx[T any](ctx context.Context, pool *Pool, fn func(context.Context, pggen.Querier) (T, error)) (T, error) {
	var result T
	err := pool.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		v, err := fn(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to invoke func: %w", err)
		}

		result = v

		return nil
	})
	if err != nil {
		return result, fmt.Errorf("failed to invoke func: %w", err)
	}

	return result, nil
}

// Query obtains a connection for the pool, executes the given function, and
// returns the connection to the pool.
func (db *Pool) Query(ctx context.Context, callback func(context.Context, pggen.Querier) error) error {
	if conn, ok := fromContext(ctx); ok {
		querier, err := pggen.NewQuerier(ctx, conn)
		if err != nil {
			return fmt.Errorf("failed to construct querier with ctx conn: %w", err)
		}

		err = callback(ctx, querier)
		if err != nil {
			return fmt.Errorf("failed to invoke func with ctx conn: %w", err)
		}

		return nil
	}

	err := db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		querier, err := pggen.NewQuerier(ctx, c.Conn())
		if err != nil {
			return fmt.Errorf("failed to consturct querier from pool: %w", err)
		}

		err = callback(ctx, querier)
		if err != nil {
			return fmt.Errorf("failed to invoke func with pool conn: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to invoke func: %w", err)
	}

	return nil
}

// Tx provides the caller with a callback in which all operations are conducted
// within a transaction.
func (db *Pool) Tx(ctx context.Context, callback func(context.Context, pggen.Querier) error) error {
	var conn interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	} = db.Pool

	// Use connection from context if found
	if ctxConn, ok := fromContext(ctx); ok {
		conn = ctxConn
	}

	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		querier, err := pggen.NewQuerier(ctx, tx.Conn())
		if err != nil {
			return fmt.Errorf("failed to construct querier from tx conn: %w", err)
		}

		ctx = newContext(ctx, tx.Conn())
		return callback(ctx, querier)
	})
}

// Exec acquires a connection from the pool and executes the given SQL. If the
// context contains a transaction then that is used.
func (db *Pool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if conn, ok := fromContext(ctx); ok {
		return conn.Exec(ctx, sql, args...)
	}
	return db.Pool.Exec(ctx, sql, args...)
}

func (db *Pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn, ok := fromContext(ctx); ok {
		return conn.QueryRow(ctx, sql, args...)
	}
	return db.Pool.QueryRow(ctx, sql, args...)
}

func (db *Pool) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	if conn, ok := fromContext(ctx); ok {
		return conn.SendBatch(ctx, b)
	}
	return db.Pool.SendBatch(ctx, b)
}

// WaitAndLock obtains an exclusive session-level advisory lock. If another
// session holds the lock with the given id then it'll wait until the other
// session releases the lock. The given fn is called once the lock is obtained
// and when the fn finishes the lock is released.
func (db *Pool) WaitAndLock(ctx context.Context, id int64, fn func(context.Context) error) (err error) {
	// A dedicated connection is obtained. Using a connection pool would cause
	// problems because a lock must be released on the same connection on which
	// it was obtained.
	return db.Pool.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
		if _, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", id); err != nil {
			return err
		}
		defer func() {
			_, closeErr := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", id)
			if err != nil {
				db.logger.Error("unlocking session-level advisory lock", "err", err)
				return
			}
			err = closeErr
		}()
		ctx = newContext(ctx, conn.Conn())
		return fn(ctx)
	})
}

func (db *Pool) Lock(ctx context.Context, table string, fn func(context.Context, pggen.Querier) error) error {
	var conn interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	} = db.Pool

	// Use connection from context if found
	if ctxConn, ok := fromContext(ctx); ok {
		conn = ctxConn
	}

	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		querier, err := pggen.NewQuerier(ctx, tx.Conn())
		if err != nil {
			return fmt.Errorf("failed to construct querier from tx conn: %w", err)
		}

		ctx = newContext(ctx, tx.Conn())

		sql := fmt.Sprintf("LOCK TABLE %s IN EXCLUSIVE MODE", table)
		if _, err := tx.Exec(ctx, sql); err != nil {
			return err
		}

		return fn(ctx, querier)
	})
}

func setDefaultMaxConnections(connString string, max int) (string, error) {
	// pg connection string can be either a URL or a DSN
	if strings.HasPrefix(connString, "postgres://") || strings.HasPrefix(connString, "postgresql://") {
		u, err := url.Parse(connString)
		if err != nil {
			return "", fmt.Errorf("parsing connection string url: %w", err)
		}
		q := u.Query()
		q.Add("pool_max_conns", strconv.Itoa(max))
		u.RawQuery = q.Encode()
		return url.PathUnescape(u.String())
	} else if connString == "" {
		// presume empty DSN
		return fmt.Sprintf("pool_max_conns=%d", max), nil
	} else {
		// presume non-empty DSN
		return fmt.Sprintf("%s pool_max_conns=%d", connString, max), nil
	}
}
