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

const insertNotificationConfigurationSQL = `INSERT INTO notification_configurations (
    notification_configuration_id,
    created_at,
    updated_at,
    name,
    url,
    triggers,
    destination_type,
    enabled,
    workspace_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
)
;`

type InsertNotificationConfigurationParams struct {
	NotificationConfigurationID pgtype.Text        `json:"notification_configuration_id"`
	CreatedAt                   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt                   pgtype.Timestamptz `json:"updated_at"`
	Name                        pgtype.Text        `json:"name"`
	URL                         pgtype.Text        `json:"url"`
	Triggers                    []string           `json:"triggers"`
	DestinationType             pgtype.Text        `json:"destination_type"`
	Enabled                     pgtype.Bool        `json:"enabled"`
	WorkspaceID                 pgtype.Text        `json:"workspace_id"`
}

// InsertNotificationConfiguration implements Querier.InsertNotificationConfiguration.
func (q *DBQuerier) InsertNotificationConfiguration(ctx context.Context, params InsertNotificationConfigurationParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertNotificationConfiguration")
	cmdTag, err := q.conn.Exec(ctx, insertNotificationConfigurationSQL, params.NotificationConfigurationID, params.CreatedAt, params.UpdatedAt, params.Name, params.URL, params.Triggers, params.DestinationType, params.Enabled, params.WorkspaceID)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("exec query InsertNotificationConfiguration: %w", err)
	}
	return cmdTag, err
}

const findNotificationConfigurationsByWorkspaceIDSQL = `SELECT *
FROM notification_configurations
WHERE workspace_id = $1
;`

type FindNotificationConfigurationsByWorkspaceIDRow struct {
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

// FindNotificationConfigurationsByWorkspaceID implements Querier.FindNotificationConfigurationsByWorkspaceID.
func (q *DBQuerier) FindNotificationConfigurationsByWorkspaceID(ctx context.Context, workspaceID pgtype.Text) ([]FindNotificationConfigurationsByWorkspaceIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindNotificationConfigurationsByWorkspaceID")
	rows, err := q.conn.Query(ctx, findNotificationConfigurationsByWorkspaceIDSQL, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("query FindNotificationConfigurationsByWorkspaceID: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindNotificationConfigurationsByWorkspaceIDRow, error) {
		var item FindNotificationConfigurationsByWorkspaceIDRow
		if err := row.Scan(&item.NotificationConfigurationID, // 'notification_configuration_id', 'NotificationConfigurationID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,       // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,       // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,            // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.URL,             // 'url', 'URL', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Triggers,        // 'triggers', 'Triggers', '[]string', '', '[]string'
			&item.DestinationType, // 'destination_type', 'DestinationType', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.WorkspaceID,     // 'workspace_id', 'WorkspaceID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Enabled,         // 'enabled', 'Enabled', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findAllNotificationConfigurationsSQL = `SELECT *
FROM notification_configurations
;`

type FindAllNotificationConfigurationsRow struct {
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

// FindAllNotificationConfigurations implements Querier.FindAllNotificationConfigurations.
func (q *DBQuerier) FindAllNotificationConfigurations(ctx context.Context) ([]FindAllNotificationConfigurationsRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindAllNotificationConfigurations")
	rows, err := q.conn.Query(ctx, findAllNotificationConfigurationsSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindAllNotificationConfigurations: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (FindAllNotificationConfigurationsRow, error) {
		var item FindAllNotificationConfigurationsRow
		if err := row.Scan(&item.NotificationConfigurationID, // 'notification_configuration_id', 'NotificationConfigurationID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,       // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,       // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,            // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.URL,             // 'url', 'URL', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Triggers,        // 'triggers', 'Triggers', '[]string', '', '[]string'
			&item.DestinationType, // 'destination_type', 'DestinationType', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.WorkspaceID,     // 'workspace_id', 'WorkspaceID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Enabled,         // 'enabled', 'Enabled', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findNotificationConfigurationSQL = `SELECT *
FROM notification_configurations
WHERE notification_configuration_id = $1
;`

type FindNotificationConfigurationRow struct {
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

// FindNotificationConfiguration implements Querier.FindNotificationConfiguration.
func (q *DBQuerier) FindNotificationConfiguration(ctx context.Context, notificationConfigurationID pgtype.Text) (FindNotificationConfigurationRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindNotificationConfiguration")
	rows, err := q.conn.Query(ctx, findNotificationConfigurationSQL, notificationConfigurationID)
	if err != nil {
		return FindNotificationConfigurationRow{}, fmt.Errorf("query FindNotificationConfiguration: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindNotificationConfigurationRow, error) {
		var item FindNotificationConfigurationRow
		if err := row.Scan(&item.NotificationConfigurationID, // 'notification_configuration_id', 'NotificationConfigurationID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,       // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,       // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,            // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.URL,             // 'url', 'URL', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Triggers,        // 'triggers', 'Triggers', '[]string', '', '[]string'
			&item.DestinationType, // 'destination_type', 'DestinationType', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.WorkspaceID,     // 'workspace_id', 'WorkspaceID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Enabled,         // 'enabled', 'Enabled', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const findNotificationConfigurationForUpdateSQL = `SELECT *
FROM notification_configurations
WHERE notification_configuration_id = $1
FOR UPDATE
;`

type FindNotificationConfigurationForUpdateRow struct {
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

// FindNotificationConfigurationForUpdate implements Querier.FindNotificationConfigurationForUpdate.
func (q *DBQuerier) FindNotificationConfigurationForUpdate(ctx context.Context, notificationConfigurationID pgtype.Text) (FindNotificationConfigurationForUpdateRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindNotificationConfigurationForUpdate")
	rows, err := q.conn.Query(ctx, findNotificationConfigurationForUpdateSQL, notificationConfigurationID)
	if err != nil {
		return FindNotificationConfigurationForUpdateRow{}, fmt.Errorf("query FindNotificationConfigurationForUpdate: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (FindNotificationConfigurationForUpdateRow, error) {
		var item FindNotificationConfigurationForUpdateRow
		if err := row.Scan(&item.NotificationConfigurationID, // 'notification_configuration_id', 'NotificationConfigurationID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.CreatedAt,       // 'created_at', 'CreatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.UpdatedAt,       // 'updated_at', 'UpdatedAt', 'pgtype.Timestamptz', 'github.com/jackc/pgx/v5/pgtype', 'Timestamptz'
			&item.Name,            // 'name', 'Name', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.URL,             // 'url', 'URL', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Triggers,        // 'triggers', 'Triggers', '[]string', '', '[]string'
			&item.DestinationType, // 'destination_type', 'DestinationType', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.WorkspaceID,     // 'workspace_id', 'WorkspaceID', 'pgtype.Text', 'github.com/jackc/pgx/v5/pgtype', 'Text'
			&item.Enabled,         // 'enabled', 'Enabled', 'pgtype.Bool', 'github.com/jackc/pgx/v5/pgtype', 'Bool'
		); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const updateNotificationConfigurationByIDSQL = `UPDATE notification_configurations
SET
    updated_at = $1,
    enabled    = $2,
    name       = $3,
    triggers   = $4,
    url        = $5
WHERE notification_configuration_id = $6
RETURNING notification_configuration_id
;`

type UpdateNotificationConfigurationByIDParams struct {
	UpdatedAt                   pgtype.Timestamptz `json:"updated_at"`
	Enabled                     pgtype.Bool        `json:"enabled"`
	Name                        pgtype.Text        `json:"name"`
	Triggers                    []string           `json:"triggers"`
	URL                         pgtype.Text        `json:"url"`
	NotificationConfigurationID pgtype.Text        `json:"notification_configuration_id"`
}

// UpdateNotificationConfigurationByID implements Querier.UpdateNotificationConfigurationByID.
func (q *DBQuerier) UpdateNotificationConfigurationByID(ctx context.Context, params UpdateNotificationConfigurationByIDParams) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateNotificationConfigurationByID")
	rows, err := q.conn.Query(ctx, updateNotificationConfigurationByIDSQL, params.UpdatedAt, params.Enabled, params.Name, params.Triggers, params.URL, params.NotificationConfigurationID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query UpdateNotificationConfigurationByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteNotificationConfigurationByIDSQL = `DELETE FROM notification_configurations
WHERE notification_configuration_id = $1
RETURNING notification_configuration_id
;`

// DeleteNotificationConfigurationByID implements Querier.DeleteNotificationConfigurationByID.
func (q *DBQuerier) DeleteNotificationConfigurationByID(ctx context.Context, notificationConfigurationID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteNotificationConfigurationByID")
	rows, err := q.conn.Query(ctx, deleteNotificationConfigurationByIDSQL, notificationConfigurationID)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("query DeleteNotificationConfigurationByID: %w", err)
	}

	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}
