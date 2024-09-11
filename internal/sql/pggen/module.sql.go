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

const insertModuleSQL = `INSERT INTO modules (
    module_id,
    created_at,
    updated_at,
    name,
    provider,
    status,
    organization_name
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
);`

type InsertModuleParams struct {
	ID               pgtype.Text        `json:"id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	Name             pgtype.Text        `json:"name"`
	Provider         pgtype.Text        `json:"provider"`
	Status           pgtype.Text        `json:"status"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

// InsertModule implements Querier.InsertModule.
func (q *DBQuerier) InsertModule(ctx context.Context, params InsertModuleParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertModule")
	cmdTag, err := q.conn.Exec(ctx, insertModuleSQL, params.ID, params.CreatedAt, params.UpdatedAt, params.Name, params.Provider, params.Status, params.OrganizationName)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertModule: %w", err)
	}
	return cmdTag, err
}

const insertModuleVersionSQL = `INSERT INTO module_versions (
    module_version_id,
    version,
    created_at,
    updated_at,
    module_id,
    status
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;`

type InsertModuleVersionParams struct {
	ModuleVersionID pgtype.Text        `json:"module_version_id"`
	Version         pgtype.Text        `json:"version"`
	CreatedAt       pgtype.Timestamptz `json:"created_at"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
	ModuleID        pgtype.Text        `json:"module_id"`
	Status          pgtype.Text        `json:"status"`
}

type InsertModuleVersionRow struct {
	ModuleVersionID pgtype.Text        `json:"module_version_id"`
	Version         pgtype.Text        `json:"version"`
	CreatedAt       pgtype.Timestamptz `json:"created_at"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
	Status          pgtype.Text        `json:"status"`
	StatusError     pgtype.Text        `json:"status_error"`
	ModuleID        pgtype.Text        `json:"module_id"`
}

// InsertModuleVersion implements Querier.InsertModuleVersion.
func (q *DBQuerier) InsertModuleVersion(ctx context.Context, params InsertModuleVersionParams) (InsertModuleVersionRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertModuleVersion")
	rows, err := q.conn.Query(ctx, insertModuleVersionSQL, params.ModuleVersionID, params.Version, params.CreatedAt, params.UpdatedAt, params.ModuleID, params.Status)
	if err != nil {
		return InsertModuleVersionRow{}, fmt.Errorf("query InsertModuleVersion: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (InsertModuleVersionRow, error) {
		var item InsertModuleVersionRow
		if err := row.Scan(&item.ModuleVersionID, // 'module_version_id', 'ModuleVersionID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Version,     // 'version', 'Version', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,   // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,   // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Status,      // 'status', 'Status', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.StatusError, // 'status_error', 'StatusError', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.ModuleID,    // 'module_id', 'ModuleID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const listModulesByOrganizationSQL = `SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    (r.*)::"repo_connections" AS module_connection,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.organization_name = $1
;`

type ListModulesByOrganizationRow struct {
	ModuleID         pgtype.Text        `json:"module_id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	Name             pgtype.Text        `json:"name"`
	Provider         pgtype.Text        `json:"provider"`
	Status           pgtype.Text        `json:"status"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	ModuleConnection *RepoConnections   `json:"module_connection"`
	Versions         []*ModuleVersions  `json:"versions"`
}

// ListModulesByOrganization implements Querier.ListModulesByOrganization.
func (q *DBQuerier) ListModulesByOrganization(ctx context.Context, organizationName pgtype.Text) ([]ListModulesByOrganizationRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "ListModulesByOrganization")
	rows, err := q.conn.Query(ctx, listModulesByOrganizationSQL, organizationName)
	if err != nil {
		return nil, fmt.Errorf("query ListModulesByOrganization: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (ListModulesByOrganizationRow, error) {
		var item ListModulesByOrganizationRow
		if err := row.Scan(&item.ModuleID, // 'module_id', 'ModuleID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,        // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Provider,         // 'provider', 'Provider', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Status,           // 'status', 'Status', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.ModuleConnection, // 'module_connection', 'ModuleConnection', '*RepoConnections', '', '*RepoConnections'
			&item.Versions,         // 'versions', 'Versions', '[]*ModuleVersions', '', '[]*ModuleVersions'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findModuleByNameSQL = `SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    (r.*)::"repo_connections" AS module_connection,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.organization_name = $1
AND   m.name = $2
AND   m.provider = $3
;`

type FindModuleByNameParams struct {
	OrganizationName pgtype.Text `json:"organization_name"`
	Name             pgtype.Text `json:"name"`
	Provider         pgtype.Text `json:"provider"`
}

type FindModuleByNameRow struct {
	ModuleID         pgtype.Text        `json:"module_id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	Name             pgtype.Text        `json:"name"`
	Provider         pgtype.Text        `json:"provider"`
	Status           pgtype.Text        `json:"status"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	ModuleConnection *RepoConnections   `json:"module_connection"`
	Versions         []*ModuleVersions  `json:"versions"`
}

// FindModuleByName implements Querier.FindModuleByName.
func (q *DBQuerier) FindModuleByName(ctx context.Context, params FindModuleByNameParams) (FindModuleByNameRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindModuleByName")
	rows, err := q.conn.Query(ctx, findModuleByNameSQL, params.OrganizationName, params.Name, params.Provider)
	if err != nil {
		return FindModuleByNameRow{}, fmt.Errorf("query FindModuleByName: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindModuleByNameRow, error) {
		var item FindModuleByNameRow
		if err := row.Scan(&item.ModuleID, // 'module_id', 'ModuleID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,        // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Provider,         // 'provider', 'Provider', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Status,           // 'status', 'Status', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.ModuleConnection, // 'module_connection', 'ModuleConnection', '*RepoConnections', '', '*RepoConnections'
			&item.Versions,         // 'versions', 'Versions', '[]*ModuleVersions', '', '[]*ModuleVersions'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findModuleByIDSQL = `SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    (r.*)::"repo_connections" AS module_connection,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.module_id = $1
;`

type FindModuleByIDRow struct {
	ModuleID         pgtype.Text        `json:"module_id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	Name             pgtype.Text        `json:"name"`
	Provider         pgtype.Text        `json:"provider"`
	Status           pgtype.Text        `json:"status"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	ModuleConnection *RepoConnections   `json:"module_connection"`
	Versions         []*ModuleVersions  `json:"versions"`
}

// FindModuleByID implements Querier.FindModuleByID.
func (q *DBQuerier) FindModuleByID(ctx context.Context, id pgtype.Text) (FindModuleByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindModuleByID")
	rows, err := q.conn.Query(ctx, findModuleByIDSQL, id)
	if err != nil {
		return FindModuleByIDRow{}, fmt.Errorf("query FindModuleByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindModuleByIDRow, error) {
		var item FindModuleByIDRow
		if err := row.Scan(&item.ModuleID, // 'module_id', 'ModuleID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,        // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Provider,         // 'provider', 'Provider', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Status,           // 'status', 'Status', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.ModuleConnection, // 'module_connection', 'ModuleConnection', '*RepoConnections', '', '*RepoConnections'
			&item.Versions,         // 'versions', 'Versions', '[]*ModuleVersions', '', '[]*ModuleVersions'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findModuleByConnectionSQL = `SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    (r.*)::"repo_connections" AS module_connection,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
JOIN repo_connections r USING (module_id)
WHERE r.vcs_provider_id = $1
AND   r.repo_path = $2
;`

type FindModuleByConnectionRow struct {
	ModuleID         pgtype.Text        `json:"module_id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	Name             pgtype.Text        `json:"name"`
	Provider         pgtype.Text        `json:"provider"`
	Status           pgtype.Text        `json:"status"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	ModuleConnection *RepoConnections   `json:"module_connection"`
	Versions         []*ModuleVersions  `json:"versions"`
}

// FindModuleByConnection implements Querier.FindModuleByConnection.
func (q *DBQuerier) FindModuleByConnection(ctx context.Context, vcsProviderID pgtype.Text, repoPath pgtype.Text) (FindModuleByConnectionRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindModuleByConnection")
	rows, err := q.conn.Query(ctx, findModuleByConnectionSQL, vcsProviderID, repoPath)
	if err != nil {
		return FindModuleByConnectionRow{}, fmt.Errorf("query FindModuleByConnection: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindModuleByConnectionRow, error) {
		var item FindModuleByConnectionRow
		if err := row.Scan(&item.ModuleID, // 'module_id', 'ModuleID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,        // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Provider,         // 'provider', 'Provider', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Status,           // 'status', 'Status', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.ModuleConnection, // 'module_connection', 'ModuleConnection', '*RepoConnections', '', '*RepoConnections'
			&item.Versions,         // 'versions', 'Versions', '[]*ModuleVersions', '', '[]*ModuleVersions'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findModuleByModuleVersionIDSQL = `SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    (r.*)::"repo_connections" AS module_connection,
    (
        SELECT array_agg(v.*) AS versions
        FROM module_versions v
        WHERE v.module_id = m.module_id
    ) AS versions
FROM modules m
JOIN module_versions mv USING (module_id)
LEFT JOIN repo_connections r USING (module_id)
WHERE mv.module_version_id = $1
;`

type FindModuleByModuleVersionIDRow struct {
	ModuleID         pgtype.Text        `json:"module_id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	Name             pgtype.Text        `json:"name"`
	Provider         pgtype.Text        `json:"provider"`
	Status           pgtype.Text        `json:"status"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	ModuleConnection *RepoConnections   `json:"module_connection"`
	Versions         []*ModuleVersions  `json:"versions"`
}

// FindModuleByModuleVersionID implements Querier.FindModuleByModuleVersionID.
func (q *DBQuerier) FindModuleByModuleVersionID(ctx context.Context, moduleVersionID pgtype.Text) (FindModuleByModuleVersionIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindModuleByModuleVersionID")
	rows, err := q.conn.Query(ctx, findModuleByModuleVersionIDSQL, moduleVersionID)
	if err != nil {
		return FindModuleByModuleVersionIDRow{}, fmt.Errorf("query FindModuleByModuleVersionID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindModuleByModuleVersionIDRow, error) {
		var item FindModuleByModuleVersionIDRow
		if err := row.Scan(&item.ModuleID, // 'module_id', 'ModuleID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,        // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Provider,         // 'provider', 'Provider', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Status,           // 'status', 'Status', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.ModuleConnection, // 'module_connection', 'ModuleConnection', '*RepoConnections', '', '*RepoConnections'
			&item.Versions,         // 'versions', 'Versions', '[]*ModuleVersions', '', '[]*ModuleVersions'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const updateModuleStatusByIDSQL = `UPDATE modules
SET status = $1
WHERE module_id = $2
RETURNING module_id
;`

// UpdateModuleStatusByID implements Querier.UpdateModuleStatusByID.
func (q *DBQuerier) UpdateModuleStatusByID(ctx context.Context, status pgtype.Text, moduleID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateModuleStatusByID")
	rows, err := q.conn.Query(ctx, updateModuleStatusByIDSQL, status, moduleID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query UpdateModuleStatusByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const insertModuleTarballSQL = `INSERT INTO module_tarballs (
    tarball,
    module_version_id
) VALUES (
    $1,
    $2
)
RETURNING module_version_id;`

// InsertModuleTarball implements Querier.InsertModuleTarball.
func (q *DBQuerier) InsertModuleTarball(ctx context.Context, tarball []byte, moduleVersionID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertModuleTarball")
	rows, err := q.conn.Query(ctx, insertModuleTarballSQL, tarball, moduleVersionID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query InsertModuleTarball: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findModuleTarballSQL = `SELECT tarball
FROM module_tarballs
WHERE module_version_id = $1
;`

// FindModuleTarball implements Querier.FindModuleTarball.
func (q *DBQuerier) FindModuleTarball(ctx context.Context, moduleVersionID pgtype.Text) ([]byte, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindModuleTarball")
	rows, err := q.conn.Query(ctx, findModuleTarballSQL, moduleVersionID)
	if err != nil {
		return nil, fmt.Errorf("query FindModuleTarball: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) ([]byte, error) {
		var item []byte
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const updateModuleVersionStatusByIDSQL = `UPDATE module_versions
SET
    status = $1,
    status_error = $2
WHERE module_version_id = $3
RETURNING *
;`

type UpdateModuleVersionStatusByIDParams struct {
	Status          pgtype.Text `json:"status"`
	StatusError     pgtype.Text `json:"status_error"`
	ModuleVersionID pgtype.Text `json:"module_version_id"`
}

type UpdateModuleVersionStatusByIDRow struct {
	ModuleVersionID pgtype.Text        `json:"module_version_id"`
	Version         pgtype.Text        `json:"version"`
	CreatedAt       pgtype.Timestamptz `json:"created_at"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
	Status          pgtype.Text        `json:"status"`
	StatusError     pgtype.Text        `json:"status_error"`
	ModuleID        pgtype.Text        `json:"module_id"`
}

// UpdateModuleVersionStatusByID implements Querier.UpdateModuleVersionStatusByID.
func (q *DBQuerier) UpdateModuleVersionStatusByID(ctx context.Context, params UpdateModuleVersionStatusByIDParams) (UpdateModuleVersionStatusByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateModuleVersionStatusByID")
	rows, err := q.conn.Query(ctx, updateModuleVersionStatusByIDSQL, params.Status, params.StatusError, params.ModuleVersionID)
	if err != nil {
		return UpdateModuleVersionStatusByIDRow{}, fmt.Errorf("query UpdateModuleVersionStatusByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (UpdateModuleVersionStatusByIDRow, error) {
		var item UpdateModuleVersionStatusByIDRow
		if err := row.Scan(&item.ModuleVersionID, // 'module_version_id', 'ModuleVersionID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Version,     // 'version', 'Version', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,   // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,   // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Status,      // 'status', 'Status', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.StatusError, // 'status_error', 'StatusError', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.ModuleID,    // 'module_id', 'ModuleID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteModuleByIDSQL = `DELETE
FROM modules
WHERE module_id = $1
RETURNING module_id
;`

// DeleteModuleByID implements Querier.DeleteModuleByID.
func (q *DBQuerier) DeleteModuleByID(ctx context.Context, moduleID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteModuleByID")
	rows, err := q.conn.Query(ctx, deleteModuleByIDSQL, moduleID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query DeleteModuleByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteModuleVersionByIDSQL = `DELETE
FROM module_versions
WHERE module_version_id = $1
RETURNING module_version_id
;`

// DeleteModuleVersionByID implements Querier.DeleteModuleVersionByID.
func (q *DBQuerier) DeleteModuleVersionByID(ctx context.Context, moduleVersionID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteModuleVersionByID")
	rows, err := q.conn.Query(ctx, deleteModuleVersionByIDSQL, moduleVersionID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query DeleteModuleVersionByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}
