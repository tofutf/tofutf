// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ genericConn = (*pgx.Conn)(nil)
var _ RegisterConn = (*pgx.Conn)(nil)

const insertAgentPoolSQL = `INSERT INTO agent_pools (
    agent_pool_id,
    name,
    created_at,
    organization_name,
    organization_scoped
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);`

type InsertAgentPoolParams struct {
	AgentPoolID        pgtype.Text        `json:"agent_pool_id"`
	Name               pgtype.Text        `json:"name"`
	CreatedAt          pgtype.Timestamptz `json:"created_at"`
	OrganizationName   pgtype.Text        `json:"organization_name"`
	OrganizationScoped pgtype.Bool        `json:"organization_scoped"`
}

// InsertAgentPool implements Querier.InsertAgentPool.
func (q *DBQuerier) InsertAgentPool(ctx context.Context, params InsertAgentPoolParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertAgentPool")
	cmdTag, err := q.conn.Exec(ctx, insertAgentPoolSQL, params.AgentPoolID, params.Name, params.CreatedAt, params.OrganizationName, params.OrganizationScoped)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertAgentPool: %w", err)
	}
	return cmdTag, err
}

const findAgentPoolsSQL = `SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
ORDER BY ap.created_at DESC
;`

type FindAgentPoolsRow struct {
	AgentPoolID         pgtype.Text        `json:"agent_pool_id"`
	Name                pgtype.Text        `json:"name"`
	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	OrganizationName    pgtype.Text        `json:"organization_name"`
	OrganizationScoped  pgtype.Bool        `json:"organization_scoped"`
	WorkspaceIds        []string           `json:"workspace_ids"`
	AllowedWorkspaceIds []string           `json:"allowed_workspace_ids"`
}

// FindAgentPools implements Querier.FindAgentPools.
func (q *DBQuerier) FindAgentPools(ctx context.Context) ([]FindAgentPoolsRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindAgentPools")
	rows, err := q.conn.Query(ctx, findAgentPoolsSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindAgentPools: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindAgentPoolsRow, error) {
		var item FindAgentPoolsRow
		if err := row.Scan(&item.AgentPoolID, // 'agent_pool_id', 'AgentPoolID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Name,                // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,           // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.OrganizationName,    // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationScoped,  // 'organization_scoped', 'OrganizationScoped', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.WorkspaceIds,        // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
			&item.AllowedWorkspaceIds, // 'allowed_workspace_ids', 'AllowedWorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findAgentPoolsByOrganizationSQL = `SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
LEFT JOIN (agent_pool_allowed_workspaces aw JOIN workspaces w USING (workspace_id)) ON ap.agent_pool_id = aw.agent_pool_id
WHERE ap.organization_name = $1
AND   (($2::text IS NULL) OR ap.name LIKE '%' || $2 || '%')
AND   (($3::text IS NULL) OR
       ap.organization_scoped OR
       w.name = $3
      )
AND   (($4::text IS NULL) OR
       ap.organization_scoped OR
       w.workspace_id = $4
      )
GROUP BY ap.agent_pool_id
ORDER BY ap.created_at DESC
;`

type FindAgentPoolsByOrganizationParams struct {
	OrganizationName     pgtype.Text `json:"organization_name"`
	NameSubstring        pgtype.Text `json:"name_substring"`
	AllowedWorkspaceName pgtype.Text `json:"allowed_workspace_name"`
	AllowedWorkspaceID   pgtype.Text `json:"allowed_workspace_id"`
}

type FindAgentPoolsByOrganizationRow struct {
	AgentPoolID         pgtype.Text        `json:"agent_pool_id"`
	Name                pgtype.Text        `json:"name"`
	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	OrganizationName    pgtype.Text        `json:"organization_name"`
	OrganizationScoped  pgtype.Bool        `json:"organization_scoped"`
	WorkspaceIds        []string           `json:"workspace_ids"`
	AllowedWorkspaceIds []string           `json:"allowed_workspace_ids"`
}

// FindAgentPoolsByOrganization implements Querier.FindAgentPoolsByOrganization.
func (q *DBQuerier) FindAgentPoolsByOrganization(ctx context.Context, params FindAgentPoolsByOrganizationParams) ([]FindAgentPoolsByOrganizationRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindAgentPoolsByOrganization")
	rows, err := q.conn.Query(ctx, findAgentPoolsByOrganizationSQL, params.OrganizationName, params.NameSubstring, params.AllowedWorkspaceName, params.AllowedWorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("query FindAgentPoolsByOrganization: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindAgentPoolsByOrganizationRow, error) {
		var item FindAgentPoolsByOrganizationRow
		if err := row.Scan(&item.AgentPoolID, // 'agent_pool_id', 'AgentPoolID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Name,                // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,           // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.OrganizationName,    // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationScoped,  // 'organization_scoped', 'OrganizationScoped', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.WorkspaceIds,        // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
			&item.AllowedWorkspaceIds, // 'allowed_workspace_ids', 'AllowedWorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findAgentPoolSQL = `SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
WHERE ap.agent_pool_id = $1
GROUP BY ap.agent_pool_id
;`

type FindAgentPoolRow struct {
	AgentPoolID         pgtype.Text        `json:"agent_pool_id"`
	Name                pgtype.Text        `json:"name"`
	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	OrganizationName    pgtype.Text        `json:"organization_name"`
	OrganizationScoped  pgtype.Bool        `json:"organization_scoped"`
	WorkspaceIds        []string           `json:"workspace_ids"`
	AllowedWorkspaceIds []string           `json:"allowed_workspace_ids"`
}

// FindAgentPool implements Querier.FindAgentPool.
func (q *DBQuerier) FindAgentPool(ctx context.Context, poolID pgtype.Text) (FindAgentPoolRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindAgentPool")
	rows, err := q.conn.Query(ctx, findAgentPoolSQL, poolID)
	if err != nil {
		return FindAgentPoolRow{}, fmt.Errorf("query FindAgentPool: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindAgentPoolRow, error) {
		var item FindAgentPoolRow
		if err := row.Scan(&item.AgentPoolID, // 'agent_pool_id', 'AgentPoolID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Name,                // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,           // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.OrganizationName,    // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationScoped,  // 'organization_scoped', 'OrganizationScoped', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.WorkspaceIds,        // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
			&item.AllowedWorkspaceIds, // 'allowed_workspace_ids', 'AllowedWorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findAgentPoolByAgentTokenIDSQL = `SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
JOIN agent_tokens at USING (agent_pool_id)
WHERE at.agent_token_id = $1
GROUP BY ap.agent_pool_id
;`

type FindAgentPoolByAgentTokenIDRow struct {
	AgentPoolID         pgtype.Text        `json:"agent_pool_id"`
	Name                pgtype.Text        `json:"name"`
	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	OrganizationName    pgtype.Text        `json:"organization_name"`
	OrganizationScoped  pgtype.Bool        `json:"organization_scoped"`
	WorkspaceIds        []string           `json:"workspace_ids"`
	AllowedWorkspaceIds []string           `json:"allowed_workspace_ids"`
}

// FindAgentPoolByAgentTokenID implements Querier.FindAgentPoolByAgentTokenID.
func (q *DBQuerier) FindAgentPoolByAgentTokenID(ctx context.Context, agentTokenID pgtype.Text) (FindAgentPoolByAgentTokenIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindAgentPoolByAgentTokenID")
	rows, err := q.conn.Query(ctx, findAgentPoolByAgentTokenIDSQL, agentTokenID)
	if err != nil {
		return FindAgentPoolByAgentTokenIDRow{}, fmt.Errorf("query FindAgentPoolByAgentTokenID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindAgentPoolByAgentTokenIDRow, error) {
		var item FindAgentPoolByAgentTokenIDRow
		if err := row.Scan(&item.AgentPoolID, // 'agent_pool_id', 'AgentPoolID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Name,                // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,           // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.OrganizationName,    // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationScoped,  // 'organization_scoped', 'OrganizationScoped', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.WorkspaceIds,        // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
			&item.AllowedWorkspaceIds, // 'allowed_workspace_ids', 'AllowedWorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const updateAgentPoolSQL = `UPDATE agent_pools
SET name = $1,
    organization_scoped = $2
WHERE agent_pool_id = $3
RETURNING *;`

type UpdateAgentPoolParams struct {
	Name               pgtype.Text `json:"name"`
	OrganizationScoped pgtype.Bool `json:"organization_scoped"`
	PoolID             pgtype.Text `json:"pool_id"`
}

type UpdateAgentPoolRow struct {
	AgentPoolID        pgtype.Text        `json:"agent_pool_id"`
	Name               pgtype.Text        `json:"name"`
	CreatedAt          pgtype.Timestamptz `json:"created_at"`
	OrganizationName   pgtype.Text        `json:"organization_name"`
	OrganizationScoped pgtype.Bool        `json:"organization_scoped"`
}

// UpdateAgentPool implements Querier.UpdateAgentPool.
func (q *DBQuerier) UpdateAgentPool(ctx context.Context, params UpdateAgentPoolParams) (UpdateAgentPoolRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateAgentPool")
	rows, err := q.conn.Query(ctx, updateAgentPoolSQL, params.Name, params.OrganizationScoped, params.PoolID)
	if err != nil {
		return UpdateAgentPoolRow{}, fmt.Errorf("query UpdateAgentPool: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (UpdateAgentPoolRow, error) {
		var item UpdateAgentPoolRow
		if err := row.Scan(&item.AgentPoolID, // 'agent_pool_id', 'AgentPoolID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Name,               // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,          // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.OrganizationName,   // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationScoped, // 'organization_scoped', 'OrganizationScoped', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteAgentPoolSQL = `DELETE
FROM agent_pools
WHERE agent_pool_id = $1
RETURNING *
;`

type DeleteAgentPoolRow struct {
	AgentPoolID        pgtype.Text        `json:"agent_pool_id"`
	Name               pgtype.Text        `json:"name"`
	CreatedAt          pgtype.Timestamptz `json:"created_at"`
	OrganizationName   pgtype.Text        `json:"organization_name"`
	OrganizationScoped pgtype.Bool        `json:"organization_scoped"`
}

// DeleteAgentPool implements Querier.DeleteAgentPool.
func (q *DBQuerier) DeleteAgentPool(ctx context.Context, poolID pgtype.Text) (DeleteAgentPoolRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteAgentPool")
	rows, err := q.conn.Query(ctx, deleteAgentPoolSQL, poolID)
	if err != nil {
		return DeleteAgentPoolRow{}, fmt.Errorf("query DeleteAgentPool: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (DeleteAgentPoolRow, error) {
		var item DeleteAgentPoolRow
		if err := row.Scan(&item.AgentPoolID, // 'agent_pool_id', 'AgentPoolID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Name,               // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,          // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.OrganizationName,   // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationScoped, // 'organization_scoped', 'OrganizationScoped', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const insertAgentPoolAllowedWorkspaceSQL = `INSERT INTO agent_pool_allowed_workspaces (
    agent_pool_id,
    workspace_id
) VALUES (
    $1,
    $2
);`

// InsertAgentPoolAllowedWorkspace implements Querier.InsertAgentPoolAllowedWorkspace.
func (q *DBQuerier) InsertAgentPoolAllowedWorkspace(ctx context.Context, poolID pgtype.Text, workspaceID pgtype.Text) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertAgentPoolAllowedWorkspace")
	cmdTag, err := q.conn.Exec(ctx, insertAgentPoolAllowedWorkspaceSQL, poolID, workspaceID)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertAgentPoolAllowedWorkspace: %w", err)
	}
	return cmdTag, err
}

const deleteAgentPoolAllowedWorkspaceSQL = `DELETE
FROM agent_pool_allowed_workspaces
WHERE agent_pool_id = $1
AND workspace_id = $2
;`

// DeleteAgentPoolAllowedWorkspace implements Querier.DeleteAgentPoolAllowedWorkspace.
func (q *DBQuerier) DeleteAgentPoolAllowedWorkspace(ctx context.Context, poolID pgtype.Text, workspaceID pgtype.Text) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteAgentPoolAllowedWorkspace")
	cmdTag, err := q.conn.Exec(ctx, deleteAgentPoolAllowedWorkspaceSQL, poolID, workspaceID)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query DeleteAgentPoolAllowedWorkspace: %w", err)
	}
	return cmdTag, err
}
