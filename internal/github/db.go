package github

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type (
	// pgdb is a github app database on postgres
	pgdb struct {
		*sql.Pool // provides access to generated SQL queries
	}

	// row represents a database row for a github app
	row struct {
		GithubAppID   pgtype.Int8 `json:"github_app_id"`
		WebhookSecret pgtype.Text `json:"webhook_secret"`
		PrivateKey    pgtype.Text `json:"private_key"`
		Slug          pgtype.Text `json:"slug"`
		Organization  pgtype.Text `json:"organization"`
	}
)

func (r row) convert() *App {
	app := &App{
		ID:            r.GithubAppID.Int64,
		Slug:          r.Slug.String,
		WebhookSecret: r.WebhookSecret.String,
		PrivateKey:    r.PrivateKey.String,
	}
	if r.Organization.Valid {
		app.Organization = &r.Organization.String
	}
	return app
}

func (db *pgdb) create(ctx context.Context, app *App) error {
	return db.Func(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertGithubApp(ctx, pggen.InsertGithubAppParams{
			GithubAppID:   pgtype.Int8{Int64: app.ID, Valid: true},
			WebhookSecret: sql.String(app.WebhookSecret),
			PrivateKey:    sql.String(app.PrivateKey),
			Slug:          sql.String(app.Slug),
			Organization:  sql.StringPtr(app.Organization),
		})
		return err
	})
}

func (db *pgdb) get(ctx context.Context) (*App, error) {
	return sql.Func(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*App, error) {
		result, err := q.FindGithubApp(ctx)
		if err != nil {
			return nil, sql.Error(err)
		}

		return row(result).convert(), nil
	})
}

func (db *pgdb) delete(ctx context.Context) error {
	return db.Lock(ctx, "github_apps", func(ctx context.Context, q pggen.Querier) error {
		result, err := q.FindGithubApp(ctx)
		if err != nil {
			return sql.Error(err)
		}

		_, err = q.DeleteGithubApp(ctx, result.GithubAppID)
		if err != nil {
			return sql.Error(err)
		}
		return nil
	})
}
