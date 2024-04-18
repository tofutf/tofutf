package connections

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type (
	db struct {
		*sql.Pool
	}
)

func (db *db) createConnection(ctx context.Context, opts ConnectOptions) error {
	return db.Func(ctx, func(ctx context.Context, q pggen.Querier) error {
		params := pggen.InsertRepoConnectionParams{
			VCSProviderID: sql.String(opts.VCSProviderID),
			RepoPath:      sql.String(opts.RepoPath),
		}

		switch opts.ConnectionType {
		case WorkspaceConnection:
			params.WorkspaceID = sql.String(opts.ResourceID)
			params.ModuleID = pgtype.Text{Valid: false}
		case ModuleConnection:
			params.ModuleID = sql.String(opts.ResourceID)
			params.WorkspaceID = pgtype.Text{Valid: false}
		default:
			return fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
		}

		if _, err := q.InsertRepoConnection(ctx, params); err != nil {
			return sql.Error(err)
		}

		return nil
	})
}

func (db *db) deleteConnection(ctx context.Context, opts DisconnectOptions) (err error) {
	return db.Func(ctx, func(ctx context.Context, q pggen.Querier) error {
		switch opts.ConnectionType {
		case WorkspaceConnection:
			_, err = q.DeleteWorkspaceConnectionByID(ctx, sql.String(opts.ResourceID))
		case ModuleConnection:
			_, err = q.DeleteModuleConnectionByID(ctx, sql.String(opts.ResourceID))
		default:
			return fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
		}
		return err
	})
}
