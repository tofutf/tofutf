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

const insertVariableSetSQL = `INSERT INTO variable_sets (
    variable_set_id,
    global,
    name,
    description,
    organization_name
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);`

type InsertVariableSetParams struct {
	VariableSetID    pgtype.Text `json:"variable_set_id"`
	Global           pgtype.Bool `json:"global"`
	Name             pgtype.Text `json:"name"`
	Description      pgtype.Text `json:"description"`
	OrganizationName pgtype.Text `json:"organization_name"`
}

// InsertVariableSet implements Querier.InsertVariableSet.
func (q *DBQuerier) InsertVariableSet(ctx context.Context, params InsertVariableSetParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertVariableSet")
	cmdTag, err := q.conn.Exec(ctx, insertVariableSetSQL, params.VariableSetID, params.Global, params.Name, params.Description, params.OrganizationName)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertVariableSet: %w", err)
	}
	return cmdTag, err
}

const findVariableSetsByOrganizationSQL = `SELECT
    *,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE organization_name = $1;`

type FindVariableSetsByOrganizationRow struct {
	VariableSetID    pgtype.Text  `json:"variable_set_id"`
	Global           pgtype.Bool  `json:"global"`
	Name             pgtype.Text  `json:"name"`
	Description      pgtype.Text  `json:"description"`
	OrganizationName pgtype.Text  `json:"organization_name"`
	Variables        []*Variables `json:"variables"`
	WorkspaceIds     []string     `json:"workspace_ids"`
}

// FindVariableSetsByOrganization implements Querier.FindVariableSetsByOrganization.
func (q *DBQuerier) FindVariableSetsByOrganization(ctx context.Context, organizationName pgtype.Text) ([]FindVariableSetsByOrganizationRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVariableSetsByOrganization")
	rows, err := q.conn.Query(ctx, findVariableSetsByOrganizationSQL, organizationName)
	if err != nil {
		return nil, fmt.Errorf("query FindVariableSetsByOrganization: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindVariableSetsByOrganizationRow, error) {
		var item FindVariableSetsByOrganizationRow
		if err := row.Scan(&item.VariableSetID, // 'variable_set_id', 'VariableSetID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Global,           // 'global', 'Global', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Description,      // 'description', 'Description', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Variables,        // 'variables', 'Variables', '[]*Variables', '', '[]*Variables'
			&item.WorkspaceIds,     // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findVariableSetsByWorkspaceSQL = `SELECT
    vs.*,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN variable_set_workspaces vsw USING (variable_set_id)
WHERE workspace_id = $1
UNION
SELECT
    vs.*,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN (organizations o JOIN workspaces w ON o.name = w.organization_name) ON o.name = vs.organization_name
WHERE vs.global IS true
AND w.workspace_id = $1;`

type FindVariableSetsByWorkspaceRow struct {
	VariableSetID    pgtype.Text  `json:"variable_set_id"`
	Global           pgtype.Bool  `json:"global"`
	Name             pgtype.Text  `json:"name"`
	Description      pgtype.Text  `json:"description"`
	OrganizationName pgtype.Text  `json:"organization_name"`
	Variables        []*Variables `json:"variables"`
	WorkspaceIds     []string     `json:"workspace_ids"`
}

// FindVariableSetsByWorkspace implements Querier.FindVariableSetsByWorkspace.
func (q *DBQuerier) FindVariableSetsByWorkspace(ctx context.Context, workspaceID pgtype.Text) ([]FindVariableSetsByWorkspaceRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVariableSetsByWorkspace")
	rows, err := q.conn.Query(ctx, findVariableSetsByWorkspaceSQL, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("query FindVariableSetsByWorkspace: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindVariableSetsByWorkspaceRow, error) {
		var item FindVariableSetsByWorkspaceRow
		if err := row.Scan(&item.VariableSetID, // 'variable_set_id', 'VariableSetID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Global,           // 'global', 'Global', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Description,      // 'description', 'Description', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Variables,        // 'variables', 'Variables', '[]*Variables', '', '[]*Variables'
			&item.WorkspaceIds,     // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findVariableSetBySetIDSQL = `SELECT
    *,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE vs.variable_set_id = $1;`

type FindVariableSetBySetIDRow struct {
	VariableSetID    pgtype.Text  `json:"variable_set_id"`
	Global           pgtype.Bool  `json:"global"`
	Name             pgtype.Text  `json:"name"`
	Description      pgtype.Text  `json:"description"`
	OrganizationName pgtype.Text  `json:"organization_name"`
	Variables        []*Variables `json:"variables"`
	WorkspaceIds     []string     `json:"workspace_ids"`
}

// FindVariableSetBySetID implements Querier.FindVariableSetBySetID.
func (q *DBQuerier) FindVariableSetBySetID(ctx context.Context, variableSetID pgtype.Text) (FindVariableSetBySetIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVariableSetBySetID")
	rows, err := q.conn.Query(ctx, findVariableSetBySetIDSQL, variableSetID)
	if err != nil {
		return FindVariableSetBySetIDRow{}, fmt.Errorf("query FindVariableSetBySetID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindVariableSetBySetIDRow, error) {
		var item FindVariableSetBySetIDRow
		if err := row.Scan(&item.VariableSetID, // 'variable_set_id', 'VariableSetID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Global,           // 'global', 'Global', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Description,      // 'description', 'Description', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Variables,        // 'variables', 'Variables', '[]*Variables', '', '[]*Variables'
			&item.WorkspaceIds,     // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findVariableSetByVariableIDSQL = `SELECT
    vs.*,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN variable_set_variables vsv USING (variable_set_id)
WHERE vsv.variable_id = $1;`

type FindVariableSetByVariableIDRow struct {
	VariableSetID    pgtype.Text  `json:"variable_set_id"`
	Global           pgtype.Bool  `json:"global"`
	Name             pgtype.Text  `json:"name"`
	Description      pgtype.Text  `json:"description"`
	OrganizationName pgtype.Text  `json:"organization_name"`
	Variables        []*Variables `json:"variables"`
	WorkspaceIds     []string     `json:"workspace_ids"`
}

// FindVariableSetByVariableID implements Querier.FindVariableSetByVariableID.
func (q *DBQuerier) FindVariableSetByVariableID(ctx context.Context, variableID pgtype.Text) (FindVariableSetByVariableIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVariableSetByVariableID")
	rows, err := q.conn.Query(ctx, findVariableSetByVariableIDSQL, variableID)
	if err != nil {
		return FindVariableSetByVariableIDRow{}, fmt.Errorf("query FindVariableSetByVariableID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindVariableSetByVariableIDRow, error) {
		var item FindVariableSetByVariableIDRow
		if err := row.Scan(&item.VariableSetID, // 'variable_set_id', 'VariableSetID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Global,           // 'global', 'Global', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Description,      // 'description', 'Description', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Variables,        // 'variables', 'Variables', '[]*Variables', '', '[]*Variables'
			&item.WorkspaceIds,     // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findVariableSetForUpdateSQL = `SELECT
    *,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE variable_set_id = $1
FOR UPDATE OF vs;`

type FindVariableSetForUpdateRow struct {
	VariableSetID    pgtype.Text  `json:"variable_set_id"`
	Global           pgtype.Bool  `json:"global"`
	Name             pgtype.Text  `json:"name"`
	Description      pgtype.Text  `json:"description"`
	OrganizationName pgtype.Text  `json:"organization_name"`
	Variables        []*Variables `json:"variables"`
	WorkspaceIds     []string     `json:"workspace_ids"`
}

// FindVariableSetForUpdate implements Querier.FindVariableSetForUpdate.
func (q *DBQuerier) FindVariableSetForUpdate(ctx context.Context, variableSetID pgtype.Text) (FindVariableSetForUpdateRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVariableSetForUpdate")
	rows, err := q.conn.Query(ctx, findVariableSetForUpdateSQL, variableSetID)
	if err != nil {
		return FindVariableSetForUpdateRow{}, fmt.Errorf("query FindVariableSetForUpdate: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindVariableSetForUpdateRow, error) {
		var item FindVariableSetForUpdateRow
		if err := row.Scan(&item.VariableSetID, // 'variable_set_id', 'VariableSetID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Global,           // 'global', 'Global', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Description,      // 'description', 'Description', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Variables,        // 'variables', 'Variables', '[]*Variables', '', '[]*Variables'
			&item.WorkspaceIds,     // 'workspace_ids', 'WorkspaceIds', '[]string', '', '[]string'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const updateVariableSetByIDSQL = `UPDATE variable_sets
SET
    global = $1,
    name = $2,
    description = $3
WHERE variable_set_id = $4
RETURNING variable_set_id;`

type UpdateVariableSetByIDParams struct {
	Global        pgtype.Bool `json:"global"`
	Name          pgtype.Text `json:"name"`
	Description   pgtype.Text `json:"description"`
	VariableSetID pgtype.Text `json:"variable_set_id"`
}

// UpdateVariableSetByID implements Querier.UpdateVariableSetByID.
func (q *DBQuerier) UpdateVariableSetByID(ctx context.Context, params UpdateVariableSetByIDParams) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateVariableSetByID")
	rows, err := q.conn.Query(ctx, updateVariableSetByIDSQL, params.Global, params.Name, params.Description, params.VariableSetID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query UpdateVariableSetByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteVariableSetByIDSQL = `DELETE
FROM variable_sets
WHERE variable_set_id = $1
RETURNING *;`

type DeleteVariableSetByIDRow struct {
	VariableSetID    pgtype.Text `json:"variable_set_id"`
	Global           pgtype.Bool `json:"global"`
	Name             pgtype.Text `json:"name"`
	Description      pgtype.Text `json:"description"`
	OrganizationName pgtype.Text `json:"organization_name"`
}

// DeleteVariableSetByID implements Querier.DeleteVariableSetByID.
func (q *DBQuerier) DeleteVariableSetByID(ctx context.Context, variableSetID pgtype.Text) (DeleteVariableSetByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteVariableSetByID")
	rows, err := q.conn.Query(ctx, deleteVariableSetByIDSQL, variableSetID)
	if err != nil {
		return DeleteVariableSetByIDRow{}, fmt.Errorf("query DeleteVariableSetByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (DeleteVariableSetByIDRow, error) {
		var item DeleteVariableSetByIDRow
		if err := row.Scan(&item.VariableSetID, // 'variable_set_id', 'VariableSetID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Global,           // 'global', 'Global', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Description,      // 'description', 'Description', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const insertVariableSetVariableSQL = `INSERT INTO variable_set_variables (
    variable_set_id,
    variable_id
) VALUES (
    $1,
    $2
);`

// InsertVariableSetVariable implements Querier.InsertVariableSetVariable.
func (q *DBQuerier) InsertVariableSetVariable(ctx context.Context, variableSetID pgtype.Text, variableID pgtype.Text) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertVariableSetVariable")
	cmdTag, err := q.conn.Exec(ctx, insertVariableSetVariableSQL, variableSetID, variableID)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertVariableSetVariable: %w", err)
	}
	return cmdTag, err
}

const deleteVariableSetVariableSQL = `DELETE
FROM variable_set_variables
WHERE variable_set_id = $1
AND variable_id = $2
RETURNING *;`

type DeleteVariableSetVariableRow struct {
	VariableSetID pgtype.Text `json:"variable_set_id"`
	VariableID    pgtype.Text `json:"variable_id"`
}

// DeleteVariableSetVariable implements Querier.DeleteVariableSetVariable.
func (q *DBQuerier) DeleteVariableSetVariable(ctx context.Context, variableSetID pgtype.Text, variableID pgtype.Text) (DeleteVariableSetVariableRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteVariableSetVariable")
	rows, err := q.conn.Query(ctx, deleteVariableSetVariableSQL, variableSetID, variableID)
	if err != nil {
		return DeleteVariableSetVariableRow{}, fmt.Errorf("query DeleteVariableSetVariable: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (DeleteVariableSetVariableRow, error) {
		var item DeleteVariableSetVariableRow
		if err := row.Scan(&item.VariableSetID, // 'variable_set_id', 'VariableSetID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VariableID, // 'variable_id', 'VariableID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const insertVariableSetWorkspaceSQL = `INSERT INTO variable_set_workspaces (
    variable_set_id,
    workspace_id
) VALUES (
    $1,
    $2
);`

// InsertVariableSetWorkspace implements Querier.InsertVariableSetWorkspace.
func (q *DBQuerier) InsertVariableSetWorkspace(ctx context.Context, variableSetID pgtype.Text, workspaceID pgtype.Text) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertVariableSetWorkspace")
	cmdTag, err := q.conn.Exec(ctx, insertVariableSetWorkspaceSQL, variableSetID, workspaceID)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertVariableSetWorkspace: %w", err)
	}
	return cmdTag, err
}

const deleteVariableSetWorkspaceSQL = `DELETE
FROM variable_set_workspaces
WHERE variable_set_id = $1
AND workspace_id = $2
RETURNING *;`

type DeleteVariableSetWorkspaceRow struct {
	VariableSetID pgtype.Text `json:"variable_set_id"`
	WorkspaceID   pgtype.Text `json:"workspace_id"`
}

// DeleteVariableSetWorkspace implements Querier.DeleteVariableSetWorkspace.
func (q *DBQuerier) DeleteVariableSetWorkspace(ctx context.Context, variableSetID pgtype.Text, workspaceID pgtype.Text) (DeleteVariableSetWorkspaceRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteVariableSetWorkspace")
	rows, err := q.conn.Query(ctx, deleteVariableSetWorkspaceSQL, variableSetID, workspaceID)
	if err != nil {
		return DeleteVariableSetWorkspaceRow{}, fmt.Errorf("query DeleteVariableSetWorkspace: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (DeleteVariableSetWorkspaceRow, error) {
		var item DeleteVariableSetWorkspaceRow
		if err := row.Scan(&item.VariableSetID, // 'variable_set_id', 'VariableSetID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.WorkspaceID, // 'workspace_id', 'WorkspaceID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteVariableSetWorkspacesSQL = `DELETE
FROM variable_set_workspaces
WHERE variable_set_id = $1;`

// DeleteVariableSetWorkspaces implements Querier.DeleteVariableSetWorkspaces.
func (q *DBQuerier) DeleteVariableSetWorkspaces(ctx context.Context, variableSetID pgtype.Text) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteVariableSetWorkspaces")
	cmdTag, err := q.conn.Exec(ctx, deleteVariableSetWorkspacesSQL, variableSetID)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query DeleteVariableSetWorkspaces: %w", err)
	}
	return cmdTag, err
}
