package repohooks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
	"github.com/tofutf/tofutf/internal/vcs"
)

type (
	db struct {
		*sql.Pool
		*internal.HostnameService
	}

	hookRow struct {
		RepohookID    pgtype.UUID `json:"repohook_id"`
		VCSID         pgtype.Text `json:"vcs_id"`
		VCSProviderID pgtype.Text `json:"vcs_provider_id"`
		Secret        pgtype.Text `json:"secret"`
		RepoPath      pgtype.Text `json:"repo_path"`
		VCSKind       pgtype.Text `json:"vcs_kind"`
	}
)

// getOrCreateHook gets a hook if it exists or creates it if it does not. Should be
// called within a tx to avoid concurrent access causing unpredictible results.
func (db *db) getOrCreateHook(ctx context.Context, h *hook) (*hook, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*hook, error) {
		result, err := q.FindRepohookByRepoAndProvider(ctx, sql.String(h.repoPath), sql.String(h.vcsProviderID))
		if err != nil {
			return nil, sql.Error(err)
		}
		if len(result) > 0 {
			return db.fromRow(hookRow(result[0]))
		}

		// not found; create instead

		insertResult, err := q.InsertRepohook(ctx, pggen.InsertRepohookParams{
			RepohookID:    sql.UUID(h.id),
			Secret:        sql.String(h.secret),
			RepoPath:      sql.String(h.repoPath),
			VCSID:         sql.StringPtr(h.cloudID),
			VCSProviderID: sql.String(h.vcsProviderID),
		})
		if err != nil {
			return nil, fmt.Errorf("inserting webhook into db: %w", sql.Error(err))
		}
		return db.fromRow(hookRow(insertResult))
	})
}

func (db *db) getHookByID(ctx context.Context, id uuid.UUID) (*hook, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*hook, error) {
		result, err := q.FindRepohookByID(ctx, sql.UUID(id))
		if err != nil {
			return nil, sql.Error(err)
		}

		return db.fromRow(hookRow(result))
	})
}

func (db *db) listHooks(ctx context.Context) ([]*hook, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*hook, error) {
		result, err := q.FindRepohooks(ctx)
		if err != nil {
			return nil, sql.Error(err)
		}

		hooks := make([]*hook, len(result))
		for i, row := range result {
			hook, err := db.fromRow(hookRow(row))
			if err != nil {
				return nil, sql.Error(err)
			}
			hooks[i] = hook
		}

		return hooks, nil
	})

}

func (db *db) listUnreferencedRepohooks(ctx context.Context) ([]*hook, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*hook, error) {
		result, err := q.FindUnreferencedRepohooks(ctx)
		if err != nil {
			return nil, sql.Error(err)
		}

		hooks := make([]*hook, len(result))
		for i, row := range result {
			hook, err := db.fromRow(hookRow(row))
			if err != nil {
				return nil, sql.Error(err)
			}

			hooks[i] = hook
		}

		return hooks, nil
	})
}

func (db *db) updateHookCloudID(ctx context.Context, id uuid.UUID, cloudID string) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.UpdateRepohookVCSID(ctx, sql.String(cloudID), sql.UUID(id))
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}

func (db *db) deleteHook(ctx context.Context, id uuid.UUID) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteRepohookByID(ctx, sql.UUID(id))
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}

// fromRow creates a hook from a database row
func (db *db) fromRow(row hookRow) (*hook, error) {
	opts := newRepohookOptions{
		id:              internal.UUID(row.RepohookID.Bytes),
		vcsProviderID:   row.VCSProviderID.String,
		secret:          internal.String(row.Secret.String),
		repoPath:        row.RepoPath.String,
		cloud:           vcs.Kind(row.VCSKind.String),
		HostnameService: db.HostnameService,
	}
	if row.VCSID.Valid {
		opts.cloudID = internal.String(row.VCSID.String)
	}

	return newRepohook(opts)
}
