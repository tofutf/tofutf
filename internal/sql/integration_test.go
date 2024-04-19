package sql_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

func TestPool(t *testing.T) {
	pg := sql.NewTestContainer(t)
	ctx := context.Background()

	t.Run("Tx", func(t *testing.T) {
		t.Run("should commit changes when think passes", func(t *testing.T) {
			pool := sql.TestContainerReset(t, pg)

			err := pool.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.InsertAgent(ctx, pggen.InsertAgentParams{
					AgentID:      sql.String("id"),
					Name:         sql.String("name"),
					Version:      sql.String("v1.0.0"),
					MaxJobs:      sql.Int4(3),
					IPAddress:    net.IPv4(192, 168, 1, 100),
					LastPingAt:   sql.Timestamptz(time.Now()),
					LastStatusAt: sql.Timestamptz(time.Now()),
					Status:       sql.String("idle"),
					AgentPoolID:  sql.NullString(),
				})
				return err
			})
			require.NoError(t, err)

			// the previous transaction should have been committed.
			err = pool.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.FindAgentByID(ctx, sql.String("id"))
				return err
			})
			require.NoError(t, err)
		})

		t.Run("should rollback changes when thunk fails", func(t *testing.T) {
			pool := sql.TestContainerReset(t, pg)

			err := pool.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.InsertAgent(ctx, pggen.InsertAgentParams{
					AgentID:      sql.String("id"),
					Name:         sql.String("name"),
					Version:      sql.String("v1.0.0"),
					MaxJobs:      sql.Int4(3),
					IPAddress:    net.IPv4(192, 168, 1, 100),
					LastPingAt:   sql.Timestamptz(time.Now()),
					LastStatusAt: sql.Timestamptz(time.Now()),
					Status:       sql.String("idle"),
					AgentPoolID:  sql.NullString(),
				})
				require.NoError(t, err)

				_, err = q.FindAgentByID(ctx, sql.String("id"))
				require.NoError(t, err)

				return fmt.Errorf("fake error")
			})
			require.Error(t, err)

			// the previous transaction should have been rolled back.
			err = pool.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.FindAgentByID(ctx, sql.String("id"))
				return err
			})
			require.Error(t, err)
		})

		t.Run("should not allow changes made from inside a tx to appear outside of the tx", func(t *testing.T) {
			pool := sql.TestContainerReset(t, pg)

			ctxA := context.Background()

			err := pool.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.InsertAgent(ctx, pggen.InsertAgentParams{
					AgentID:      sql.String("id"),
					Name:         sql.String("name"),
					Version:      sql.String("v1.0.0"),
					MaxJobs:      sql.Int4(3),
					IPAddress:    net.IPv4(192, 168, 1, 100),
					LastPingAt:   sql.Timestamptz(time.Now()),
					LastStatusAt: sql.Timestamptz(time.Now()),
					Status:       sql.String("idle"),
					AgentPoolID:  sql.NullString(),
				})
				if err != nil {
					return err
				}

				_, err = q.FindAgentByID(ctx, sql.String("id"))
				if err != nil {
					return err
				}

				// the agent should not exist.
				err = pool.Query(ctxA, func(ctx context.Context, q pggen.Querier) error {
					_, err := q.FindAgentByID(ctx, sql.String("id"))
					return err
				})
				require.Error(t, err)

				return nil
			})
			require.NoError(t, err)
		})
		t.Run("invoking while already in a transaction should not result in committing the top level transaction", func(t *testing.T) {
			pool := sql.TestContainerReset(t, pg)

			ctxA := context.Background()

			err := pool.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
				err := pool.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
					_, err := q.InsertAgent(ctx, pggen.InsertAgentParams{
						AgentID:      sql.String("id"),
						Name:         sql.String("name"),
						Version:      sql.String("v1.0.0"),
						MaxJobs:      sql.Int4(3),
						IPAddress:    net.IPv4(192, 168, 1, 100),
						LastPingAt:   sql.Timestamptz(time.Now()),
						LastStatusAt: sql.Timestamptz(time.Now()),
						Status:       sql.String("idle"),
						AgentPoolID:  sql.NullString(),
					})
					require.NoError(t, err)

					_, err = q.FindAgentByID(ctx, sql.String("id"))
					require.NoError(t, err)

					return nil
				})
				require.NoError(t, err)

				// the previous transaction should not have been committed yet.
				err = pool.Query(ctxA, func(ctx context.Context, q pggen.Querier) error {
					_, err := q.FindAgentByID(ctx, sql.String("id"))
					return err
				})
				require.Error(t, err)

				return nil
			})
			require.NoError(t, err)
		})

		t.Run("should handle nested calls to Query and Tx", func(t *testing.T) {
			t.SkipNow()
			pool := sql.TestContainerReset(t, pg)

			org, err := organization.NewOrganization(organization.CreateOptions{
				Name: internal.String("acmeco"),
			})
			require.NoError(t, err)

			aCtx := context.Background()

			err = pool.Tx(ctx, func(txCtx context.Context, q pggen.Querier) error {
				_, err := q.InsertOrganization(txCtx, pggen.InsertOrganizationParams{
					ID:                         sql.String(org.ID),
					CreatedAt:                  sql.Timestamptz(org.CreatedAt),
					UpdatedAt:                  sql.Timestamptz(org.UpdatedAt),
					Name:                       sql.String(org.Name),
					Email:                      sql.StringPtr(org.Email),
					CollaboratorAuthPolicy:     sql.StringPtr(org.CollaboratorAuthPolicy),
					CostEstimationEnabled:      sql.Bool(org.CostEstimationEnabled),
					SessionRemember:            sql.Int4Ptr(org.SessionRemember),
					SessionTimeout:             sql.Int4Ptr(org.SessionTimeout),
					AllowForceDeleteWorkspaces: sql.Bool(org.AllowForceDeleteWorkspaces),
				})
				if err != nil {
					return err
				}

				// this should succeed because it is using the same querier from the
				// same tx
				_, err = q.FindOrganizationByID(txCtx, sql.String(org.ID))
				assert.NoError(t, err)

				// this should succeed because it is using the same ctx from the same tx
				err = pool.Query(txCtx, func(ctx context.Context, q pggen.Querier) error {
					_, err = q.FindOrganizationByID(ctx, sql.String(org.ID))
					return err
				})
				assert.NoError(t, err)

				err = pool.Tx(txCtx, func(ctx context.Context, q pggen.Querier) error {
					// this should succeed because it is using a child tx via the
					// querier
					_, err = q.FindOrganizationByID(ctx, sql.String(org.ID))
					assert.NoError(t, err)

					err := pool.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
						// this should succeed because it is using a child tx via the
						// context
						_, err = q.FindOrganizationByID(ctx, sql.String(org.ID))
						assert.NoError(t, err)

						return nil
					})
					assert.NoError(t, err)

					return nil
				})
				require.NoError(t, err)

				err = pool.Query(aCtx, func(ctx context.Context, q pggen.Querier) error {
					// this should fail because it is using a different ctx
					_, err := q.FindOrganizationByID(ctx, sql.String(org.ID))
					require.Error(t, err)
					assert.True(t, sql.NoRowsInResultError(err))

					return nil
				})
				require.Nil(t, err)

				return nil
			})
			require.NoError(t, err)
		})
	})

	t.Run("Query", func(t *testing.T) {
		t.Run("should invoke callback with querier", func(t *testing.T) {
			pool := sql.TestContainerReset(t, pg)

			err := pool.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
				_, err := q.InsertAgent(ctx, pggen.InsertAgentParams{
					AgentID:      sql.String("id"),
					Name:         sql.String("name"),
					Version:      sql.String("v1.0.0"),
					MaxJobs:      sql.Int4(3),
					IPAddress:    net.IPv4(192, 168, 1, 100),
					LastPingAt:   sql.Timestamptz(time.Now()),
					LastStatusAt: sql.Timestamptz(time.Now()),
					Status:       sql.String("idle"),
					AgentPoolID:  sql.NullString(),
				})
				if err != nil {
					return err
				}

				_, err = q.FindAgentByID(ctx, sql.String("id"))
				if err != nil {
					return err
				}

				return nil
			})
			require.NoError(t, err)
		})
	})

	// TestWaitAndLock tests acquiring a connection from a pool, obtaining a session
	// lock and then releasing lock and the connection, and it does this several
	// times, to demonstrate that it is returning resources and not running into
	// limits.
	t.Run("WaitAndLock", func(t *testing.T) {
		pool := sql.TestContainerReset(t, pg)

		for i := 0; i < 100; i++ {
			func() {
				err := pool.WaitAndLock(ctx, 123, func(context.Context) error { return nil })
				require.NoError(t, err)
			}()
		}
	})
}
