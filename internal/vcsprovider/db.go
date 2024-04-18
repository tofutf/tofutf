package vcsprovider

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal/github"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
	"github.com/tofutf/tofutf/internal/vcs"
)

type (
	// pgdb is a VCS provider database on postgres
	pgdb struct {
		// provides access to generated SQL queries
		*sql.Pool
		*factory
	}
	// pgrow represents a database row for a vcs provider
	pgrow struct {
		VCSProviderID    pgtype.Text             `json:"vcs_provider_id"`
		Token            pgtype.Text             `json:"token"`
		CreatedAt        pgtype.Timestamptz      `json:"created_at"`
		Name             pgtype.Text             `json:"name"`
		VCSKind          pgtype.Text             `json:"vcs_kind"`
		OrganizationName pgtype.Text             `json:"organization_name"`
		GithubAppID      pgtype.Int8             `json:"github_app_id"`
		GithubApp        pggen.GithubApps        `json:"github_app"`
		GithubAppInstall pggen.GithubAppInstalls `json:"github_app_install"`
	}
)

func (db *pgdb) create(ctx context.Context, provider *VCSProvider) error {
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		params := pggen.InsertVCSProviderParams{
			VCSProviderID:    sql.String(provider.ID),
			Name:             sql.String(provider.Name),
			VCSKind:          sql.String(string(provider.Kind)),
			OrganizationName: sql.String(provider.Organization),
			CreatedAt:        sql.Timestamptz(provider.CreatedAt),
			Token:            sql.StringPtr(provider.Token),
		}
		if provider.GithubApp != nil {
			params.GithubAppID = pgtype.Int8{Int64: provider.GithubApp.AppCredentials.ID, Valid: true}
		} else {
			params.GithubAppID = pgtype.Int8{Valid: false}
		}
		_, err := q.InsertVCSProvider(ctx, params)
		if err != nil {
			return err
		}
		if provider.GithubApp != nil {
			_, err := q.InsertGithubAppInstall(ctx, pggen.InsertGithubAppInstallParams{
				GithubAppID:   pgtype.Int8{Int64: provider.GithubApp.AppCredentials.ID, Valid: true},
				InstallID:     pgtype.Int8{Int64: provider.GithubApp.ID, Valid: true},
				Username:      sql.StringPtr(provider.GithubApp.User),
				Organization:  sql.StringPtr(provider.GithubApp.Organization),
				VCSProviderID: sql.String(provider.ID),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (db *pgdb) update(ctx context.Context, id string, fn func(*VCSProvider) error) error {
	var provider *VCSProvider
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		row, err := q.FindVCSProviderForUpdate(ctx, sql.String(id))
		if err != nil {
			return sql.Error(err)
		}
		provider, err = db.toProvider(ctx, pgrow(row))
		if err != nil {
			return err
		}
		if err := fn(provider); err != nil {
			return err
		}
		_, err = q.UpdateVCSProvider(ctx, pggen.UpdateVCSProviderParams{
			VCSProviderID: sql.String(id),
			Token:         sql.StringPtr(provider.Token),
			Name:          sql.String(provider.Name),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (db *pgdb) get(ctx context.Context, id string) (*VCSProvider, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*VCSProvider, error) {
		row, err := q.FindVCSProvider(ctx, sql.String(id))
		if err != nil {
			return nil, sql.Error(err)
		}

		return db.toProvider(ctx, pgrow(row))
	})
}

func (db *pgdb) list(ctx context.Context) ([]*VCSProvider, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*VCSProvider, error) {
		rows, err := q.FindVCSProviders(ctx)
		if err != nil {
			return nil, sql.Error(err)
		}

		providers := make([]*VCSProvider, len(rows))
		for i, r := range rows {
			provider, err := db.toProvider(ctx, pgrow(r))
			if err != nil {
				return nil, err
			}
			providers[i] = provider
		}

		return providers, nil
	})
}

func (db *pgdb) listByOrganization(ctx context.Context, organization string) ([]*VCSProvider, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*VCSProvider, error) {
		rows, err := q.FindVCSProvidersByOrganization(ctx, sql.String(organization))
		if err != nil {
			return nil, sql.Error(err)
		}

		providers := make([]*VCSProvider, len(rows))
		for i, r := range rows {
			provider, err := db.toProvider(ctx, pgrow(r))
			if err != nil {
				return nil, err
			}
			providers[i] = provider
		}

		return providers, nil
	})
}

func (db *pgdb) listByGithubAppInstall(ctx context.Context, installID int64) ([]*VCSProvider, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*VCSProvider, error) {
		rows, err := q.FindVCSProvidersByGithubAppInstallID(ctx,
			pgtype.Int8{Int64: installID, Valid: true},
		)
		if err != nil {
			return nil, sql.Error(err)
		}
		providers := make([]*VCSProvider, len(rows))
		for i, r := range rows {
			provider, err := db.toProvider(ctx, pgrow(r))
			if err != nil {
				return nil, err
			}
			providers[i] = provider
		}
		return providers, nil
	})
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteVCSProviderByID(ctx, sql.String(id))
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}

// unmarshal a vcs provider row from the database.
func (db *pgdb) toProvider(ctx context.Context, row pgrow) (*VCSProvider, error) {
	opts := CreateOptions{
		Organization: row.OrganizationName.String,
		Name:         row.Name.String,
		// GithubAppService: db.Git
	}
	if row.Token.Valid {
		opts.Token = &row.Token.String
		kind := vcs.Kind(row.VCSKind.String)
		opts.Kind = &kind
	}
	var creds *github.InstallCredentials
	if row.GithubApp != (pggen.GithubApps{}) {
		creds = &github.InstallCredentials{
			ID: row.GithubAppInstall.InstallID.Int64,
			AppCredentials: github.AppCredentials{
				ID:         row.GithubApp.GithubAppID.Int64,
				PrivateKey: row.GithubApp.PrivateKey.String,
			},
		}

		if row.GithubAppInstall.Username.Valid {
			creds.User = &row.GithubAppInstall.Username.String
		}

		if row.GithubAppInstall.Organization.Valid {
			creds.Organization = &row.GithubAppInstall.Organization.String
		}
	}
	return db.fromDB(ctx, opts, creds, row.VCSProviderID.String, row.CreatedAt.Time.UTC())
}
