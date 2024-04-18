package gpgkeys

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type (
	// pgdb is a notification configuration database on postgres
	pgdb struct {
		*sql.Pool // provides access to generated SQL queries
	}

	pgresult struct {
		ID               pgtype.Text        `json:"id"`
		OrganizationName pgtype.Text        `json:"organization_name"`
		AsciiArmor       pgtype.Text        `json:"ascii_armor"`
		KeyID            pgtype.Text        `json:"key_id"`
		CreatedAt        pgtype.Timestamptz `json:"created_at"`
		UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	}

	pgUpdateOpts struct {
		organizationName    string
		keyID               string
		newOrganizationName string
		updatedAt           time.Time
	}

	pgDeleteOpts struct {
		keyID        string
		organization string
	}

	pgGetOptions struct {
		organization string
		keyID        string
	}
)

func (r pgresult) toRegistryGPGKey() *GPGKey {
	return &GPGKey{
		OrganizationName: r.OrganizationName.String,
		ID:               r.ID.String,
		ASCIIArmor:       r.AsciiArmor.String,
		CreatedAt:        r.CreatedAt.Time,
		KeyID:            r.KeyID.String,
		UpdatedAt:        r.UpdatedAt.Time,
	}
}

type GPGKey struct {
	ID               string
	OrganizationName string
	ASCIIArmor       string
	CreatedAt        time.Time
	UpdatedAt        time.Time

	KeyID string
}

func (db pgdb) getRegistryGPGKey(ctx context.Context, opts pgGetOptions) (*GPGKey, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*GPGKey, error) {
		row, err := q.GetGPGKey(ctx, sql.String(opts.keyID), sql.String(opts.organization))
		if err != nil {
			return nil, sql.Error(err)
		}

		return pgresult(row).toRegistryGPGKey(), nil
	})
}

func (db *pgdb) listRegistryGPGKeys(ctx context.Context, organizationName []string) ([]*GPGKey, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*GPGKey, error) {
		rows, err := q.ListGPGKeys(ctx, organizationName)
		if err != nil {
			return nil, sql.Error(err)
		}

		keys := make([]*GPGKey, len(rows))
		for i, row := range rows {
			keys[i] = pgresult(row).toRegistryGPGKey()
		}

		return keys, nil
	})
}

func (db *pgdb) deleteRegistryGPGKey(ctx context.Context, opts pgDeleteOpts) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		response, err := q.DeleteGPGKey(ctx, sql.String(opts.keyID), sql.String(opts.organization))
		if err != nil {
			return sql.Error(err)
		}

		if count := response.RowsAffected(); count != 1 {
			return sql.Error(fmt.Errorf("unable to delete registry gpg key"))
		}

		return nil
	})
}

func (db *pgdb) updateRegistryGPGKey(ctx context.Context, opts pgUpdateOpts) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		response, err := q.UpdateGPGKey(ctx, pggen.UpdateGPGKeyParams{
			OrganizationName:    sql.String(opts.organizationName),
			NewOrganizationName: sql.String(opts.newOrganizationName),
			KeyID:               sql.String(opts.keyID),
			UpdatedAt:           sql.Timestamptz(opts.updatedAt),
		})
		if err != nil {
			return sql.Error(err)
		}

		if count := response.RowsAffected(); count != 1 {
			return sql.Error(fmt.Errorf("unable to delete registry gpg key"))
		}

		return nil
	})
}

func (db *pgdb) createRegistryGPGKey(ctx context.Context, key *GPGKey) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertGPGKey(ctx, pggen.InsertGPGKeyParams{
			ID:               sql.String(key.ID),
			OrganizationName: sql.String(key.OrganizationName),
			AsciiArmor:       sql.String(key.ASCIIArmor),
			CreatedAt:        sql.Timestamptz(key.CreatedAt),
			UpdatedAt:        sql.Timestamptz(key.UpdatedAt),
			KeyID:            sql.String(key.KeyID),
		})
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}
