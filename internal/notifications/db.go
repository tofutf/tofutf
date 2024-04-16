package notifications

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type (
	// pgdb is a notification configuration database on postgres
	pgdb struct {
		*sql.Pool // provides access to generated SQL queries
	}

	pgresult struct {
		NotificationConfigurationID pgtype.Text        `json:"notification_configuration_id"`
		CreatedAt                   pgtype.Timestamptz `json:"created_at"`
		UpdatedAt                   pgtype.Timestamptz `json:"updated_at"`
		Name                        pgtype.Text        `json:"name"`
		URL                         pgtype.Text        `json:"url"`
		Triggers                    []string           `json:"triggers"`
		DestinationType             pgtype.Text        `json:"destination_type"`
		WorkspaceID                 pgtype.Text        `json:"workspace_id"`
		Enabled                     pgtype.Bool        `json:"enabled"`
	}
)

func (r pgresult) toNotificationConfiguration() *Config {
	nc := &Config{
		ID:              r.NotificationConfigurationID.String,
		CreatedAt:       r.CreatedAt.Time.UTC(),
		UpdatedAt:       r.UpdatedAt.Time.UTC(),
		Name:            r.Name.String,
		Enabled:         r.Enabled.Bool,
		DestinationType: Destination(r.DestinationType.String),
		WorkspaceID:     r.WorkspaceID.String,
	}
	for _, t := range r.Triggers {
		nc.Triggers = append(nc.Triggers, Trigger(t))
	}
	if r.URL.Valid {
		nc.URL = &r.URL.String
	}
	return nc
}

func (db *pgdb) create(ctx context.Context, nc *Config) error {
	return db.Func(ctx, func(ctx context.Context, q pggen.Querier) error {
		params := pggen.InsertNotificationConfigurationParams{
			NotificationConfigurationID: sql.String(nc.ID),
			CreatedAt:                   sql.Timestamptz(nc.CreatedAt),
			UpdatedAt:                   sql.Timestamptz(nc.UpdatedAt),
			Name:                        sql.String(nc.Name),
			Enabled:                     sql.Bool(nc.Enabled),
			DestinationType:             sql.String(string(nc.DestinationType)),
			URL:                         sql.NullString(),
			WorkspaceID:                 sql.String(nc.WorkspaceID),
		}
		for _, t := range nc.Triggers {
			params.Triggers = append(params.Triggers, string(t))
		}
		if nc.URL != nil {
			params.URL = sql.String(*nc.URL)
		}

		_, err := q.InsertNotificationConfiguration(ctx, params)
		return sql.Error(err)
	})
}

func (db *pgdb) update(ctx context.Context, id string, updateFunc func(*Config) error) (*Config, error) {
	return sql.Tx(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*Config, error) {
		result, err := q.FindNotificationConfigurationForUpdate(ctx, sql.String(id))
		if err != nil {
			return nil, sql.Error(err)
		}
		nc := pgresult(result).toNotificationConfiguration()
		if err := updateFunc(nc); err != nil {
			return nil, sql.Error(err)
		}
		params := pggen.UpdateNotificationConfigurationByIDParams{
			UpdatedAt:                   sql.Timestamptz(internal.CurrentTimestamp(nil)),
			Enabled:                     sql.Bool(nc.Enabled),
			Name:                        sql.String(nc.Name),
			URL:                         sql.NullString(),
			NotificationConfigurationID: sql.String(nc.ID),
		}
		for _, t := range nc.Triggers {
			params.Triggers = append(params.Triggers, string(t))
		}
		if nc.URL != nil {
			params.URL = sql.String(*nc.URL)
		}

		_, err = q.UpdateNotificationConfigurationByID(ctx, params)
		if err != nil {
			return nil, err
		}

		return nc, nil
	})
}

func (db *pgdb) list(ctx context.Context, workspaceID string) ([]*Config, error) {
	return sql.Func(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*Config, error) {
		results, err := q.FindNotificationConfigurationsByWorkspaceID(ctx, sql.String(workspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}

		configs := make([]*Config, len(results))
		for i, row := range results {
			configs[i] = pgresult(row).toNotificationConfiguration()
		}

		return configs, nil
	})
}

func (db *pgdb) listAll(ctx context.Context) ([]*Config, error) {
	return sql.Func(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*Config, error) {
		results, err := q.FindAllNotificationConfigurations(ctx)
		if err != nil {
			return nil, sql.Error(err)
		}

		configs := make([]*Config, len(results))
		for i, row := range results {
			configs[i] = pgresult(row).toNotificationConfiguration()
		}
		return configs, nil
	})
}

func (db *pgdb) get(ctx context.Context, id string) (*Config, error) {
	return sql.Func(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*Config, error) {
		row, err := q.FindNotificationConfiguration(ctx, sql.String(id))
		if err != nil {
			return nil, sql.Error(err)
		}

		return pgresult(row).toNotificationConfiguration(), nil
	})
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	return db.Func(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteNotificationConfigurationByID(ctx, sql.String(id))
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}
