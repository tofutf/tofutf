// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ genericConn = (*pgx.Conn)(nil)
var _ RegisterConn = (*pgx.Conn)(nil)

const insertRepohookSQL = `WITH inserted AS (
    INSERT INTO repohooks (
        repohook_id,
        vcs_id,
        vcs_provider_id,
        secret,
        repo_path
    ) VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM inserted w
JOIN vcs_providers v USING (vcs_provider_id);`

type InsertRepohookParams struct {
	RepohookID    pgtype.UUID `json:"repohook_id"`
	VCSID         pgtype.Text `json:"vcs_id"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
	Secret        pgtype.Text `json:"secret"`
	RepoPath      pgtype.Text `json:"repo_path"`
}

type InsertRepohookRow struct {
	RepohookID    pgtype.UUID `json:"repohook_id"`
	VCSID         pgtype.Text `json:"vcs_id"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
	Secret        pgtype.Text `json:"secret"`
	RepoPath      pgtype.Text `json:"repo_path"`
	VCSKind       pgtype.Text `json:"vcs_kind"`
}

// InsertRepohook implements Querier.InsertRepohook.
func (q *DBQuerier) InsertRepohook(ctx context.Context, params InsertRepohookParams) (InsertRepohookRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertRepohook")
	rows, err := q.conn.Query(ctx, insertRepohookSQL, params.RepohookID, params.VCSID, params.VCSProviderID, params.Secret, params.RepoPath)
	if err != nil {
		return InsertRepohookRow{}, fmt.Errorf("query InsertRepohook: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (InsertRepohookRow, error) {
		var item InsertRepohookRow
		if err := row.Scan(&item.RepohookID, // 'repohook_id', 'RepohookID', 'pgtype.UUID', 'github.com/jackc/pgx/v5/pgtype', 'UUID'
			&item.VCSID,         // 'vcs_id', 'VCSID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Secret,        // 'secret', 'Secret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.RepoPath,      // 'repo_path', 'RepoPath', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,       // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const updateRepohookVCSIDSQL = `UPDATE repohooks
SET vcs_id = $1
WHERE repohook_id = $2
RETURNING *;`

type UpdateRepohookVCSIDRow struct {
	RepohookID    pgtype.UUID `json:"repohook_id"`
	VCSID         pgtype.Text `json:"vcs_id"`
	Secret        pgtype.Text `json:"secret"`
	RepoPath      pgtype.Text `json:"repo_path"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
}

// UpdateRepohookVCSID implements Querier.UpdateRepohookVCSID.
func (q *DBQuerier) UpdateRepohookVCSID(ctx context.Context, vcsID pgtype.Text, repohookID pgtype.UUID) (UpdateRepohookVCSIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateRepohookVCSID")
	rows, err := q.conn.Query(ctx, updateRepohookVCSIDSQL, vcsID, repohookID)
	if err != nil {
		return UpdateRepohookVCSIDRow{}, fmt.Errorf("query UpdateRepohookVCSID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (UpdateRepohookVCSIDRow, error) {
		var item UpdateRepohookVCSIDRow
		if err := row.Scan(&item.RepohookID, // 'repohook_id', 'RepohookID', 'pgtype.UUID', 'github.com/jackc/pgx/v5/pgtype', 'UUID'
			&item.VCSID,         // 'vcs_id', 'VCSID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Secret,        // 'secret', 'Secret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.RepoPath,      // 'repo_path', 'RepoPath', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findRepohooksSQL = `SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id);`

type FindRepohooksRow struct {
	RepohookID    pgtype.UUID `json:"repohook_id"`
	VCSID         pgtype.Text `json:"vcs_id"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
	Secret        pgtype.Text `json:"secret"`
	RepoPath      pgtype.Text `json:"repo_path"`
	VCSKind       pgtype.Text `json:"vcs_kind"`
}

// FindRepohooks implements Querier.FindRepohooks.
func (q *DBQuerier) FindRepohooks(ctx context.Context) ([]FindRepohooksRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindRepohooks")
	rows, err := q.conn.Query(ctx, findRepohooksSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindRepohooks: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindRepohooksRow, error) {
		var item FindRepohooksRow
		if err := row.Scan(&item.RepohookID, // 'repohook_id', 'RepohookID', 'pgtype.UUID', 'github.com/jackc/pgx/v5/pgtype', 'UUID'
			&item.VCSID,         // 'vcs_id', 'VCSID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Secret,        // 'secret', 'Secret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.RepoPath,      // 'repo_path', 'RepoPath', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,       // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findRepohookByIDSQL = `SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE w.repohook_id = $1;`

type FindRepohookByIDRow struct {
	RepohookID    pgtype.UUID `json:"repohook_id"`
	VCSID         pgtype.Text `json:"vcs_id"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
	Secret        pgtype.Text `json:"secret"`
	RepoPath      pgtype.Text `json:"repo_path"`
	VCSKind       pgtype.Text `json:"vcs_kind"`
}

// FindRepohookByID implements Querier.FindRepohookByID.
func (q *DBQuerier) FindRepohookByID(ctx context.Context, repohookID pgtype.UUID) (FindRepohookByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindRepohookByID")
	rows, err := q.conn.Query(ctx, findRepohookByIDSQL, repohookID)
	if err != nil {
		return FindRepohookByIDRow{}, fmt.Errorf("query FindRepohookByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindRepohookByIDRow, error) {
		var item FindRepohookByIDRow
		if err := row.Scan(&item.RepohookID, // 'repohook_id', 'RepohookID', 'pgtype.UUID', 'github.com/jackc/pgx/v5/pgtype', 'UUID'
			&item.VCSID,         // 'vcs_id', 'VCSID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Secret,        // 'secret', 'Secret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.RepoPath,      // 'repo_path', 'RepoPath', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,       // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findRepohookByRepoAndProviderSQL = `SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE repo_path = $1
AND   vcs_provider_id = $2;`

type FindRepohookByRepoAndProviderRow struct {
	RepohookID    pgtype.UUID `json:"repohook_id"`
	VCSID         pgtype.Text `json:"vcs_id"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
	Secret        pgtype.Text `json:"secret"`
	RepoPath      pgtype.Text `json:"repo_path"`
	VCSKind       pgtype.Text `json:"vcs_kind"`
}

// FindRepohookByRepoAndProvider implements Querier.FindRepohookByRepoAndProvider.
func (q *DBQuerier) FindRepohookByRepoAndProvider(ctx context.Context, repoPath pgtype.Text, vcsProviderID pgtype.Text) ([]FindRepohookByRepoAndProviderRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindRepohookByRepoAndProvider")
	rows, err := q.conn.Query(ctx, findRepohookByRepoAndProviderSQL, repoPath, vcsProviderID)
	if err != nil {
		return nil, fmt.Errorf("query FindRepohookByRepoAndProvider: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindRepohookByRepoAndProviderRow, error) {
		var item FindRepohookByRepoAndProviderRow
		if err := row.Scan(&item.RepohookID, // 'repohook_id', 'RepohookID', 'pgtype.UUID', 'github.com/jackc/pgx/v5/pgtype', 'UUID'
			&item.VCSID,         // 'vcs_id', 'VCSID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Secret,        // 'secret', 'Secret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.RepoPath,      // 'repo_path', 'RepoPath', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,       // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findUnreferencedRepohooksSQL = `SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE NOT EXISTS (
    SELECT FROM repo_connections rc
    WHERE rc.vcs_provider_id = w.vcs_provider_id
    AND   rc.repo_path = w.repo_path
);`

type FindUnreferencedRepohooksRow struct {
	RepohookID    pgtype.UUID `json:"repohook_id"`
	VCSID         pgtype.Text `json:"vcs_id"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
	Secret        pgtype.Text `json:"secret"`
	RepoPath      pgtype.Text `json:"repo_path"`
	VCSKind       pgtype.Text `json:"vcs_kind"`
}

// FindUnreferencedRepohooks implements Querier.FindUnreferencedRepohooks.
func (q *DBQuerier) FindUnreferencedRepohooks(ctx context.Context) ([]FindUnreferencedRepohooksRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUnreferencedRepohooks")
	rows, err := q.conn.Query(ctx, findUnreferencedRepohooksSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindUnreferencedRepohooks: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindUnreferencedRepohooksRow, error) {
		var item FindUnreferencedRepohooksRow
		if err := row.Scan(&item.RepohookID, // 'repohook_id', 'RepohookID', 'pgtype.UUID', 'github.com/jackc/pgx/v5/pgtype', 'UUID'
			&item.VCSID,         // 'vcs_id', 'VCSID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Secret,        // 'secret', 'Secret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.RepoPath,      // 'repo_path', 'RepoPath', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,       // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteRepohookByIDSQL = `DELETE
FROM repohooks
WHERE repohook_id = $1
RETURNING *;`

type DeleteRepohookByIDRow struct {
	RepohookID    pgtype.UUID `json:"repohook_id"`
	VCSID         pgtype.Text `json:"vcs_id"`
	Secret        pgtype.Text `json:"secret"`
	RepoPath      pgtype.Text `json:"repo_path"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
}

// DeleteRepohookByID implements Querier.DeleteRepohookByID.
func (q *DBQuerier) DeleteRepohookByID(ctx context.Context, repohookID pgtype.UUID) (DeleteRepohookByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteRepohookByID")
	rows, err := q.conn.Query(ctx, deleteRepohookByIDSQL, repohookID)
	if err != nil {
		return DeleteRepohookByIDRow{}, fmt.Errorf("query DeleteRepohookByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (DeleteRepohookByIDRow, error) {
		var item DeleteRepohookByIDRow
		if err := row.Scan(&item.RepohookID, // 'repohook_id', 'RepohookID', 'pgtype.UUID', 'github.com/jackc/pgx/v5/pgtype', 'UUID'
			&item.VCSID,         // 'vcs_id', 'VCSID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Secret,        // 'secret', 'Secret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.RepoPath,      // 'repo_path', 'RepoPath', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}
