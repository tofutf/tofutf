// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ genericConn = (*pgx.Conn)(nil)
var _ RegisterConn = (*pgx.Conn)(nil)

const insertTeamMembershipSQL = `WITH
    users AS (
        SELECT username
        FROM unnest($1::text[]) t(username)
    )
INSERT INTO team_memberships (username, team_id)
SELECT username, $2
FROM users
RETURNING username
;`

// InsertTeamMembership implements Querier.InsertTeamMembership.
func (q *DBQuerier) InsertTeamMembership(ctx context.Context, usernames []string, teamID pgtype.Text) ([]pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertTeamMembership")
	rows, err := q.conn.Query(ctx, insertTeamMembershipSQL, usernames, teamID)
	if err != nil {
		return nil, fmt.Errorf("query InsertTeamMembership: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}

const deleteTeamMembershipSQL = `WITH
    users AS (
        SELECT username
        FROM unnest($1::text[]) t(username)
    )
DELETE
FROM team_memberships tm
USING users
WHERE
    tm.username = users.username AND
    tm.team_id  = $2
RETURNING tm.username
;`

// DeleteTeamMembership implements Querier.DeleteTeamMembership.
func (q *DBQuerier) DeleteTeamMembership(ctx context.Context, usernames []string, teamID pgtype.Text) ([]pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteTeamMembership")
	rows, err := q.conn.Query(ctx, deleteTeamMembershipSQL, usernames, teamID)
	if err != nil {
		return nil, fmt.Errorf("query DeleteTeamMembership: %w", err)
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (pgtype.Text, error) {
		var item pgtype.Text
		if err := row.Scan(&item); err != nil {
			return item, fmt.Errorf("failed to scan: %w", err)
		}
		return item, nil
	})
}
