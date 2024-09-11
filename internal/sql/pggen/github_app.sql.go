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

const insertGithubAppSQL = `INSERT INTO github_apps (
    github_app_id,
    webhook_secret,
    private_key,
    slug,
    organization
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);`

type InsertGithubAppParams struct {
	GithubAppID   pgtype.Int8 `json:"github_app_id"`
	WebhookSecret pgtype.Text `json:"webhook_secret"`
	PrivateKey    pgtype.Text `json:"private_key"`
	Slug          pgtype.Text `json:"slug"`
	Organization  pgtype.Text `json:"organization"`
}

// InsertGithubApp implements Querier.InsertGithubApp.
func (q *DBQuerier) InsertGithubApp(ctx context.Context, params InsertGithubAppParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertGithubApp")
	cmdTag, err := q.conn.Exec(ctx, insertGithubAppSQL, params.GithubAppID, params.WebhookSecret, params.PrivateKey, params.Slug, params.Organization)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertGithubApp: %w", err)
	}
	return cmdTag, err
}

const findGithubAppSQL = `SELECT *
FROM github_apps;`

type FindGithubAppRow struct {
	GithubAppID   pgtype.Int8 `json:"github_app_id"`
	WebhookSecret pgtype.Text `json:"webhook_secret"`
	PrivateKey    pgtype.Text `json:"private_key"`
	Slug          pgtype.Text `json:"slug"`
	Organization  pgtype.Text `json:"organization"`
}

// FindGithubApp implements Querier.FindGithubApp.
func (q *DBQuerier) FindGithubApp(ctx context.Context) (FindGithubAppRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindGithubApp")
	rows, err := q.conn.Query(ctx, findGithubAppSQL)
	if err != nil {
		return FindGithubAppRow{}, fmt.Errorf("query FindGithubApp: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindGithubAppRow, error) {
		var item FindGithubAppRow
		if err := row.Scan(&item.GithubAppID, // 'github_app_id', 'GithubAppID', 'pgtype.Int8', 'github.com/jackc/pgx/v5/pgtype', 'Int8'
			&item.WebhookSecret, // 'webhook_secret', 'WebhookSecret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.PrivateKey,    // 'private_key', 'PrivateKey', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Slug,          // 'slug', 'Slug', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Organization,  // 'organization', 'Organization', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteGithubAppSQL = `DELETE
FROM github_apps
WHERE github_app_id = $1
RETURNING *;`

type DeleteGithubAppRow struct {
	GithubAppID   pgtype.Int8 `json:"github_app_id"`
	WebhookSecret pgtype.Text `json:"webhook_secret"`
	PrivateKey    pgtype.Text `json:"private_key"`
	Slug          pgtype.Text `json:"slug"`
	Organization  pgtype.Text `json:"organization"`
}

// DeleteGithubApp implements Querier.DeleteGithubApp.
func (q *DBQuerier) DeleteGithubApp(ctx context.Context, githubAppID pgtype.Int8) (DeleteGithubAppRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteGithubApp")
	rows, err := q.conn.Query(ctx, deleteGithubAppSQL, githubAppID)
	if err != nil {
		return DeleteGithubAppRow{}, fmt.Errorf("query DeleteGithubApp: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (DeleteGithubAppRow, error) {
		var item DeleteGithubAppRow
		if err := row.Scan(&item.GithubAppID, // 'github_app_id', 'GithubAppID', 'pgtype.Int8', 'github.com/jackc/pgx/v5/pgtype', 'Int8'
			&item.WebhookSecret, // 'webhook_secret', 'WebhookSecret', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.PrivateKey,    // 'private_key', 'PrivateKey', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Slug,          // 'slug', 'Slug', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Organization,  // 'organization', 'Organization', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const insertGithubAppInstallSQL = `INSERT INTO github_app_installs (
    github_app_id,
    install_id,
    username,
    organization,
    vcs_provider_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);`

type InsertGithubAppInstallParams struct {
	GithubAppID   pgtype.Int8 `json:"github_app_id"`
	InstallID     pgtype.Int8 `json:"install_id"`
	Username      pgtype.Text `json:"username"`
	Organization  pgtype.Text `json:"organization"`
	VCSProviderID pgtype.Text `json:"vcs_provider_id"`
}

// InsertGithubAppInstall implements Querier.InsertGithubAppInstall.
func (q *DBQuerier) InsertGithubAppInstall(ctx context.Context, params InsertGithubAppInstallParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertGithubAppInstall")
	cmdTag, err := q.conn.Exec(ctx, insertGithubAppInstallSQL, params.GithubAppID, params.InstallID, params.Username, params.Organization, params.VCSProviderID)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertGithubAppInstall: %w", err)
	}
	return cmdTag, err
}
