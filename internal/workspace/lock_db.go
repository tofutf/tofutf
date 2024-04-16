package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

// toggleLock toggles the workspace lock state in the DB.
func (db *pgdb) toggleLock(ctx context.Context, workspaceID string, togglefn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		// retrieve workspace
		result, err := q.FindWorkspaceByIDForUpdate(ctx, sql.String(workspaceID))
		if err != nil {
			return err
		}
		ws, err = pgresult(result).toWorkspace()
		if err != nil {
			return err
		}
		if err := togglefn(ws); err != nil {
			return err
		}
		// persist to db
		params := pggen.UpdateWorkspaceLockByIDParams{
			WorkspaceID: pgtype.Text{String: ws.ID, Valid: true},
		}
		if ws.Lock == nil {
			params.RunID = pgtype.Text{Valid: false}
			params.Username = pgtype.Text{Valid: false}
		} else if ws.Lock.LockKind == RunLock {
			params.RunID = pgtype.Text{String: ws.Lock.id, Valid: true}
			params.Username = pgtype.Text{Valid: false}
		} else if ws.Lock.LockKind == UserLock {
			params.Username = pgtype.Text{String: ws.Lock.id, Valid: true}
			params.RunID = pgtype.Text{Valid: false}
		} else {
			return ErrWorkspaceInvalidLock
		}
		_, err = q.UpdateWorkspaceLockByID(ctx, params)
		if err != nil {
			return sql.Error(err)
		}
		return nil
	})
	return ws, err
}
