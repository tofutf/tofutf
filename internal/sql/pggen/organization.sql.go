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

const insertOrganizationSQL = `INSERT INTO organizations (
    organization_id,
    created_at,
    updated_at,
    name,
    email,
    collaborator_auth_policy,
    cost_estimation_enabled,
    session_remember,
    session_timeout,
    allow_force_delete_workspaces
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
);`

type InsertOrganizationParams struct {
	ID                         pgtype.Text        `json:"id"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt                  pgtype.Timestamptz `json:"updated_at"`
	Name                       pgtype.Text        `json:"name"`
	Email                      pgtype.Text        `json:"email"`
	CollaboratorAuthPolicy     pgtype.Text        `json:"collaborator_auth_policy"`
	CostEstimationEnabled      pgtype.Bool        `json:"cost_estimation_enabled"`
	SessionRemember            pgtype.Int4        `json:"session_remember"`
	SessionTimeout             pgtype.Int4        `json:"session_timeout"`
	AllowForceDeleteWorkspaces pgtype.Bool        `json:"allow_force_delete_workspaces"`
}

// InsertOrganization implements Querier.InsertOrganization.
func (q *DBQuerier) InsertOrganization(ctx context.Context, params InsertOrganizationParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertOrganization")
	cmdTag, err := q.conn.Exec(ctx, insertOrganizationSQL, params.ID, params.CreatedAt, params.UpdatedAt, params.Name, params.Email, params.CollaboratorAuthPolicy, params.CostEstimationEnabled, params.SessionRemember, params.SessionTimeout, params.AllowForceDeleteWorkspaces)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertOrganization: %w", err)
	}
	return cmdTag, err
}

const findOrganizationNameByWorkspaceIDSQL = `SELECT organization_name
FROM workspaces
WHERE workspace_id = $1
;`

// FindOrganizationNameByWorkspaceID implements Querier.FindOrganizationNameByWorkspaceID.
func (q *DBQuerier) FindOrganizationNameByWorkspaceID(ctx context.Context, workspaceID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindOrganizationNameByWorkspaceID")
	rows, err := q.conn.Query(ctx, findOrganizationNameByWorkspaceIDSQL, workspaceID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query FindOrganizationNameByWorkspaceID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findOrganizationByNameSQL = `SELECT * FROM organizations WHERE name = $1;`

type FindOrganizationByNameRow struct {
	OrganizationID             pgtype.Text        `json:"organization_id"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt                  pgtype.Timestamptz `json:"updated_at"`
	Name                       pgtype.Text        `json:"name"`
	SessionRemember            pgtype.Int4        `json:"session_remember"`
	SessionTimeout             pgtype.Int4        `json:"session_timeout"`
	Email                      pgtype.Text        `json:"email"`
	CollaboratorAuthPolicy     pgtype.Text        `json:"collaborator_auth_policy"`
	AllowForceDeleteWorkspaces pgtype.Bool        `json:"allow_force_delete_workspaces"`
	CostEstimationEnabled      pgtype.Bool        `json:"cost_estimation_enabled"`
}

// FindOrganizationByName implements Querier.FindOrganizationByName.
func (q *DBQuerier) FindOrganizationByName(ctx context.Context, name pgtype.Text) (FindOrganizationByNameRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindOrganizationByName")
	rows, err := q.conn.Query(ctx, findOrganizationByNameSQL, name)
	if err != nil {
		return FindOrganizationByNameRow{}, fmt.Errorf("query FindOrganizationByName: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindOrganizationByNameRow, error) {
		var item FindOrganizationByNameRow
		if err := row.Scan(&item.OrganizationID, // 'organization_id', 'OrganizationID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,                  // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,                  // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,                       // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.SessionRemember,            // 'session_remember', 'SessionRemember', 'pgtype.Int4', 'github.com/jackc/pgx/v5/pgtype', 'Int4'
			&item.SessionTimeout,             // 'session_timeout', 'SessionTimeout', 'pgtype.Int4', 'github.com/jackc/pgx/v5/pgtype', 'Int4'
			&item.Email,                      // 'email', 'Email', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CollaboratorAuthPolicy,     // 'collaborator_auth_policy', 'CollaboratorAuthPolicy', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.AllowForceDeleteWorkspaces, // 'allow_force_delete_workspaces', 'AllowForceDeleteWorkspaces', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.CostEstimationEnabled,      // 'cost_estimation_enabled', 'CostEstimationEnabled', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findOrganizationByIDSQL = `SELECT * FROM organizations WHERE organization_id = $1;`

type FindOrganizationByIDRow struct {
	OrganizationID             pgtype.Text        `json:"organization_id"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt                  pgtype.Timestamptz `json:"updated_at"`
	Name                       pgtype.Text        `json:"name"`
	SessionRemember            pgtype.Int4        `json:"session_remember"`
	SessionTimeout             pgtype.Int4        `json:"session_timeout"`
	Email                      pgtype.Text        `json:"email"`
	CollaboratorAuthPolicy     pgtype.Text        `json:"collaborator_auth_policy"`
	AllowForceDeleteWorkspaces pgtype.Bool        `json:"allow_force_delete_workspaces"`
	CostEstimationEnabled      pgtype.Bool        `json:"cost_estimation_enabled"`
}

// FindOrganizationByID implements Querier.FindOrganizationByID.
func (q *DBQuerier) FindOrganizationByID(ctx context.Context, organizationID pgtype.Text) (FindOrganizationByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindOrganizationByID")
	rows, err := q.conn.Query(ctx, findOrganizationByIDSQL, organizationID)
	if err != nil {
		return FindOrganizationByIDRow{}, fmt.Errorf("query FindOrganizationByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindOrganizationByIDRow, error) {
		var item FindOrganizationByIDRow
		if err := row.Scan(&item.OrganizationID, // 'organization_id', 'OrganizationID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,                  // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,                  // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,                       // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.SessionRemember,            // 'session_remember', 'SessionRemember', 'pgtype.Int4', 'github.com/jackc/pgx/v5/pgtype', 'Int4'
			&item.SessionTimeout,             // 'session_timeout', 'SessionTimeout', 'pgtype.Int4', 'github.com/jackc/pgx/v5/pgtype', 'Int4'
			&item.Email,                      // 'email', 'Email', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CollaboratorAuthPolicy,     // 'collaborator_auth_policy', 'CollaboratorAuthPolicy', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.AllowForceDeleteWorkspaces, // 'allow_force_delete_workspaces', 'AllowForceDeleteWorkspaces', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.CostEstimationEnabled,      // 'cost_estimation_enabled', 'CostEstimationEnabled', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findOrganizationByNameForUpdateSQL = `SELECT *
FROM organizations
WHERE name = $1
FOR UPDATE
;`

type FindOrganizationByNameForUpdateRow struct {
	OrganizationID             pgtype.Text        `json:"organization_id"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt                  pgtype.Timestamptz `json:"updated_at"`
	Name                       pgtype.Text        `json:"name"`
	SessionRemember            pgtype.Int4        `json:"session_remember"`
	SessionTimeout             pgtype.Int4        `json:"session_timeout"`
	Email                      pgtype.Text        `json:"email"`
	CollaboratorAuthPolicy     pgtype.Text        `json:"collaborator_auth_policy"`
	AllowForceDeleteWorkspaces pgtype.Bool        `json:"allow_force_delete_workspaces"`
	CostEstimationEnabled      pgtype.Bool        `json:"cost_estimation_enabled"`
}

// FindOrganizationByNameForUpdate implements Querier.FindOrganizationByNameForUpdate.
func (q *DBQuerier) FindOrganizationByNameForUpdate(ctx context.Context, name pgtype.Text) (FindOrganizationByNameForUpdateRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindOrganizationByNameForUpdate")
	rows, err := q.conn.Query(ctx, findOrganizationByNameForUpdateSQL, name)
	if err != nil {
		return FindOrganizationByNameForUpdateRow{}, fmt.Errorf("query FindOrganizationByNameForUpdate: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindOrganizationByNameForUpdateRow, error) {
		var item FindOrganizationByNameForUpdateRow
		if err := row.Scan(&item.OrganizationID, // 'organization_id', 'OrganizationID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,                  // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,                  // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,                       // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.SessionRemember,            // 'session_remember', 'SessionRemember', 'pgtype.Int4', 'github.com/jackc/pgx/v5/pgtype', 'Int4'
			&item.SessionTimeout,             // 'session_timeout', 'SessionTimeout', 'pgtype.Int4', 'github.com/jackc/pgx/v5/pgtype', 'Int4'
			&item.Email,                      // 'email', 'Email', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CollaboratorAuthPolicy,     // 'collaborator_auth_policy', 'CollaboratorAuthPolicy', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.AllowForceDeleteWorkspaces, // 'allow_force_delete_workspaces', 'AllowForceDeleteWorkspaces', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.CostEstimationEnabled,      // 'cost_estimation_enabled', 'CostEstimationEnabled', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findOrganizationsSQL = `SELECT *
FROM organizations
WHERE name LIKE ANY($1)
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3
;`

type FindOrganizationsParams struct {
	Names  []string    `json:"names"`
	Limit  pgtype.Int8 `json:"limit"`
	Offset pgtype.Int8 `json:"offset"`
}

type FindOrganizationsRow struct {
	OrganizationID             pgtype.Text        `json:"organization_id"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt                  pgtype.Timestamptz `json:"updated_at"`
	Name                       pgtype.Text        `json:"name"`
	SessionRemember            pgtype.Int4        `json:"session_remember"`
	SessionTimeout             pgtype.Int4        `json:"session_timeout"`
	Email                      pgtype.Text        `json:"email"`
	CollaboratorAuthPolicy     pgtype.Text        `json:"collaborator_auth_policy"`
	AllowForceDeleteWorkspaces pgtype.Bool        `json:"allow_force_delete_workspaces"`
	CostEstimationEnabled      pgtype.Bool        `json:"cost_estimation_enabled"`
}

// FindOrganizations implements Querier.FindOrganizations.
func (q *DBQuerier) FindOrganizations(ctx context.Context, params FindOrganizationsParams) ([]FindOrganizationsRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindOrganizations")
	rows, err := q.conn.Query(ctx, findOrganizationsSQL, params.Names, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("query FindOrganizations: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindOrganizationsRow, error) {
		var item FindOrganizationsRow
		if err := row.Scan(&item.OrganizationID, // 'organization_id', 'OrganizationID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,                  // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,                  // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,                       // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.SessionRemember,            // 'session_remember', 'SessionRemember', 'pgtype.Int4', 'github.com/jackc/pgx/v5/pgtype', 'Int4'
			&item.SessionTimeout,             // 'session_timeout', 'SessionTimeout', 'pgtype.Int4', 'github.com/jackc/pgx/v5/pgtype', 'Int4'
			&item.Email,                      // 'email', 'Email', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CollaboratorAuthPolicy,     // 'collaborator_auth_policy', 'CollaboratorAuthPolicy', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.AllowForceDeleteWorkspaces, // 'allow_force_delete_workspaces', 'AllowForceDeleteWorkspaces', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
			&item.CostEstimationEnabled,      // 'cost_estimation_enabled', 'CostEstimationEnabled', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const countOrganizationsSQL = `SELECT count(*)
FROM organizations
WHERE name LIKE ANY($1)
;`

// CountOrganizations implements Querier.CountOrganizations.
func (q *DBQuerier) CountOrganizations(ctx context.Context, names []string) (pgtype.Int8, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "CountOrganizations")
	rows, err := q.conn.Query(ctx, countOrganizationsSQL, names)
	if err != nil {
		return pgtype.Int8{}, fmt.Errorf("query CountOrganizations: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Int8, error) {
		var item pgtype.Int8
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const updateOrganizationByNameSQL = `UPDATE organizations
SET
    name = $1,
    email = $2,
    collaborator_auth_policy = $3,
    cost_estimation_enabled = $4,
    session_remember = $5,
    session_timeout = $6,
    allow_force_delete_workspaces = $7,
    updated_at = $8
WHERE name = $9
RETURNING organization_id;`

type UpdateOrganizationByNameParams struct {
	NewName                    pgtype.Text        `json:"new_name"`
	Email                      pgtype.Text        `json:"email"`
	CollaboratorAuthPolicy     pgtype.Text        `json:"collaborator_auth_policy"`
	CostEstimationEnabled      pgtype.Bool        `json:"cost_estimation_enabled"`
	SessionRemember            pgtype.Int4        `json:"session_remember"`
	SessionTimeout             pgtype.Int4        `json:"session_timeout"`
	AllowForceDeleteWorkspaces pgtype.Bool        `json:"allow_force_delete_workspaces"`
	UpdatedAt                  pgtype.Timestamptz `json:"updated_at"`
	Name                       pgtype.Text        `json:"name"`
}

// UpdateOrganizationByName implements Querier.UpdateOrganizationByName.
func (q *DBQuerier) UpdateOrganizationByName(ctx context.Context, params UpdateOrganizationByNameParams) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateOrganizationByName")
	rows, err := q.conn.Query(ctx, updateOrganizationByNameSQL, params.NewName, params.Email, params.CollaboratorAuthPolicy, params.CostEstimationEnabled, params.SessionRemember, params.SessionTimeout, params.AllowForceDeleteWorkspaces, params.UpdatedAt, params.Name)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query UpdateOrganizationByName: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteOrganizationByNameSQL = `DELETE
FROM organizations
WHERE name = $1
RETURNING organization_id;`

// DeleteOrganizationByName implements Querier.DeleteOrganizationByName.
func (q *DBQuerier) DeleteOrganizationByName(ctx context.Context, name pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteOrganizationByName")
	rows, err := q.conn.Query(ctx, deleteOrganizationByNameSQL, name)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query DeleteOrganizationByName: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}
