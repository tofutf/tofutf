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

const insertVCSProviderSQL = `INSERT INTO vcs_providers (
    vcs_provider_id,
    created_at,
    name,
    vcs_kind,
    token,
    github_app_id,
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

type InsertVCSProviderParams struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	Token            pgtype.Text        `json:"token"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

// InsertVCSProvider implements Querier.InsertVCSProvider.
func (q *DBQuerier) InsertVCSProvider(ctx context.Context, params InsertVCSProviderParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertVCSProvider")
	cmdTag, err := q.conn.Exec(ctx, insertVCSProviderSQL, params.VCSProviderID, params.CreatedAt, params.Name, params.VCSKind, params.Token, params.GithubAppID, params.OrganizationName)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertVCSProvider: %w", err)
	}
	return cmdTag, err
}

const findVCSProvidersByOrganizationSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.organization_name = $1
;`

type FindVCSProvidersByOrganizationRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProvidersByOrganization implements Querier.FindVCSProvidersByOrganization.
func (q *DBQuerier) FindVCSProvidersByOrganization(ctx context.Context, organizationName pgtype.Text) ([]FindVCSProvidersByOrganizationRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProvidersByOrganization")
	rows, err := q.conn.Query(ctx, findVCSProvidersByOrganizationSQL, organizationName)
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProvidersByOrganization: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindVCSProvidersByOrganizationRow, error) {
		var item FindVCSProvidersByOrganizationRow
		if err := row.Scan(&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Token,            // 'token', 'Token', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,          // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.GithubAppID,      // 'github_app_id', 'GithubAppID', 'pgtype.Int8', 'github.com/jackc/pgx/v5/pgtype', 'Int8'
			&item.GithubApp,        // 'github_app', 'GithubApp', '*GithubApps', '', '*GithubApps'
			&item.GithubAppInstall, // 'github_app_install', 'GithubAppInstall', '*GithubAppInstalls', '', '*GithubAppInstalls'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findVCSProvidersSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
;`

type FindVCSProvidersRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProviders implements Querier.FindVCSProviders.
func (q *DBQuerier) FindVCSProviders(ctx context.Context) ([]FindVCSProvidersRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProviders")
	rows, err := q.conn.Query(ctx, findVCSProvidersSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProviders: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindVCSProvidersRow, error) {
		var item FindVCSProvidersRow
		if err := row.Scan(&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Token,            // 'token', 'Token', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,          // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.GithubAppID,      // 'github_app_id', 'GithubAppID', 'pgtype.Int8', 'github.com/jackc/pgx/v5/pgtype', 'Int8'
			&item.GithubApp,        // 'github_app', 'GithubApp', '*GithubApps', '', '*GithubApps'
			&item.GithubAppInstall, // 'github_app_install', 'GithubAppInstall', '*GithubAppInstalls', '', '*GithubAppInstalls'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findVCSProvidersByGithubAppInstallIDSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE gi.install_id = $1
;`

type FindVCSProvidersByGithubAppInstallIDRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProvidersByGithubAppInstallID implements Querier.FindVCSProvidersByGithubAppInstallID.
func (q *DBQuerier) FindVCSProvidersByGithubAppInstallID(ctx context.Context, installID pgtype.Int8) ([]FindVCSProvidersByGithubAppInstallIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProvidersByGithubAppInstallID")
	rows, err := q.conn.Query(ctx, findVCSProvidersByGithubAppInstallIDSQL, installID)
	if err != nil {
		return nil, fmt.Errorf("query FindVCSProvidersByGithubAppInstallID: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindVCSProvidersByGithubAppInstallIDRow, error) {
		var item FindVCSProvidersByGithubAppInstallIDRow
		if err := row.Scan(&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Token,            // 'token', 'Token', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,          // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.GithubAppID,      // 'github_app_id', 'GithubAppID', 'pgtype.Int8', 'github.com/jackc/pgx/v5/pgtype', 'Int8'
			&item.GithubApp,        // 'github_app', 'GithubApp', '*GithubApps', '', '*GithubApps'
			&item.GithubAppInstall, // 'github_app_install', 'GithubAppInstall', '*GithubAppInstalls', '', '*GithubAppInstalls'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findVCSProviderSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.vcs_provider_id = $1
;`

type FindVCSProviderRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProvider implements Querier.FindVCSProvider.
func (q *DBQuerier) FindVCSProvider(ctx context.Context, vcsProviderID pgtype.Text) (FindVCSProviderRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProvider")
	rows, err := q.conn.Query(ctx, findVCSProviderSQL, vcsProviderID)
	if err != nil {
		return FindVCSProviderRow{}, fmt.Errorf("query FindVCSProvider: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindVCSProviderRow, error) {
		var item FindVCSProviderRow
		if err := row.Scan(&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Token,            // 'token', 'Token', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,          // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.GithubAppID,      // 'github_app_id', 'GithubAppID', 'pgtype.Int8', 'github.com/jackc/pgx/v5/pgtype', 'Int8'
			&item.GithubApp,        // 'github_app', 'GithubApp', '*GithubApps', '', '*GithubApps'
			&item.GithubAppInstall, // 'github_app_install', 'GithubAppInstall', '*GithubAppInstalls', '', '*GithubAppInstalls'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findVCSProviderForUpdateSQL = `SELECT
    v.*,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.vcs_provider_id = $1
FOR UPDATE OF v
;`

type FindVCSProviderForUpdateRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
	GithubApp        *GithubApps        `json:"github_app"`
	GithubAppInstall *GithubAppInstalls `json:"github_app_install"`
}

// FindVCSProviderForUpdate implements Querier.FindVCSProviderForUpdate.
func (q *DBQuerier) FindVCSProviderForUpdate(ctx context.Context, vcsProviderID pgtype.Text) (FindVCSProviderForUpdateRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindVCSProviderForUpdate")
	rows, err := q.conn.Query(ctx, findVCSProviderForUpdateSQL, vcsProviderID)
	if err != nil {
		return FindVCSProviderForUpdateRow{}, fmt.Errorf("query FindVCSProviderForUpdate: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindVCSProviderForUpdateRow, error) {
		var item FindVCSProviderForUpdateRow
		if err := row.Scan(&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Token,            // 'token', 'Token', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,          // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.GithubAppID,      // 'github_app_id', 'GithubAppID', 'pgtype.Int8', 'github.com/jackc/pgx/v5/pgtype', 'Int8'
			&item.GithubApp,        // 'github_app', 'GithubApp', '*GithubApps', '', '*GithubApps'
			&item.GithubAppInstall, // 'github_app_install', 'GithubAppInstall', '*GithubAppInstalls', '', '*GithubAppInstalls'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const updateVCSProviderSQL = `UPDATE vcs_providers
SET name = $1, token = $2
WHERE vcs_provider_id = $3
RETURNING *
;`

type UpdateVCSProviderParams struct {
	Name          pgtype.Text `json:"name"`
	Token         pgtype.Text `json:"token"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
}

type UpdateVCSProviderRow struct {
	VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	VCSKind          pgtype.Text        `json:"vcs_kind"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	GithubAppID      pgtype.Int8        `json:"github_app_id"`
}

// UpdateVCSProvider implements Querier.UpdateVCSProvider.
func (q *DBQuerier) UpdateVCSProvider(ctx context.Context, params UpdateVCSProviderParams) (UpdateVCSProviderRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateVCSProvider")
	rows, err := q.conn.Query(ctx, updateVCSProviderSQL, params.Name, params.Token, params.VCSProviderID)
	if err != nil {
		return UpdateVCSProviderRow{}, fmt.Errorf("query UpdateVCSProvider: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (UpdateVCSProviderRow, error) {
		var item UpdateVCSProviderRow
		if err := row.Scan(&item.VCSProviderID, // 'vcs_provider_id', 'VCSProviderID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Token,            // 'token', 'Token', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,        // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,             // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.VCSKind,          // 'vcs_kind', 'VCSKind', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.OrganizationName, // 'organization_name', 'OrganizationName', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.GithubAppID,      // 'github_app_id', 'GithubAppID', 'pgtype.Int8', 'github.com/jackc/pgx/v5/pgtype', 'Int8'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteVCSProviderByIDSQL = `DELETE
FROM vcs_providers
WHERE vcs_provider_id = $1
RETURNING vcs_provider_id
;`

// DeleteVCSProviderByID implements Querier.DeleteVCSProviderByID.
func (q *DBQuerier) DeleteVCSProviderByID(ctx context.Context, vcsProviderID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteVCSProviderByID")
	rows, err := q.conn.Query(ctx, deleteVCSProviderByIDSQL, vcsProviderID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query DeleteVCSProviderByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}
