// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const insertUserSQL = `INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username
) VALUES (
    $1,
    $2,
    $3,
    $4
);`

type InsertUserParams struct {
	ID        pgtype.Text
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  pgtype.Text
}

// InsertUser implements Querier.InsertUser.
func (q *DBQuerier) InsertUser(ctx context.Context, params InsertUserParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertUser")
	cmdTag, err := q.conn.Exec(ctx, insertUserSQL, params.ID, params.CreatedAt, params.UpdatedAt, params.Username)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertUser: %w", err)
	}
	return cmdTag, err
}

// InsertUserBatch implements Querier.InsertUserBatch.
func (q *DBQuerier) InsertUserBatch(batch genericBatch, params InsertUserParams) {
	batch.Queue(insertUserSQL, params.ID, params.CreatedAt, params.UpdatedAt, params.Username)
}

// InsertUserScan implements Querier.InsertUserScan.
func (q *DBQuerier) InsertUserScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertUserBatch: %w", err)
	}
	return cmdTag, err
}

const findUsersSQL = `SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
GROUP BY u.user_id
;`

type FindUsersRow struct {
	UserID        pgtype.Text     `json:"user_id"`
	Username      pgtype.Text     `json:"username"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Sessions      []Sessions      `json:"sessions"`
	Tokens        []Tokens        `json:"tokens"`
	Organizations []Organizations `json:"organizations"`
}

// FindUsers implements Querier.FindUsers.
func (q *DBQuerier) FindUsers(ctx context.Context) ([]FindUsersRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUsers")
	rows, err := q.conn.Query(ctx, findUsersSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindUsers: %w", err)
	}
	defer rows.Close()
	items := []FindUsersRow{}
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	for rows.Next() {
		var item FindUsersRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsers row: %w", err)
		}
		if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
			return nil, fmt.Errorf("assign FindUsers row: %w", err)
		}
		if err := tokensArray.AssignTo(&item.Tokens); err != nil {
			return nil, fmt.Errorf("assign FindUsers row: %w", err)
		}
		if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
			return nil, fmt.Errorf("assign FindUsers row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsers rows: %w", err)
	}
	return items, err
}

// FindUsersBatch implements Querier.FindUsersBatch.
func (q *DBQuerier) FindUsersBatch(batch genericBatch) {
	batch.Queue(findUsersSQL)
}

// FindUsersScan implements Querier.FindUsersScan.
func (q *DBQuerier) FindUsersScan(results pgx.BatchResults) ([]FindUsersRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindUsersBatch: %w", err)
	}
	defer rows.Close()
	items := []FindUsersRow{}
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	for rows.Next() {
		var item FindUsersRow
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
			return nil, fmt.Errorf("scan FindUsersBatch row: %w", err)
		}
		if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
			return nil, fmt.Errorf("assign FindUsers row: %w", err)
		}
		if err := tokensArray.AssignTo(&item.Tokens); err != nil {
			return nil, fmt.Errorf("assign FindUsers row: %w", err)
		}
		if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
			return nil, fmt.Errorf("assign FindUsers row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindUsersBatch rows: %w", err)
	}
	return items, err
}

const findUserByIDSQL = `SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE u.user_id = $1
GROUP BY u.user_id
;`

type FindUserByIDRow struct {
	UserID        pgtype.Text     `json:"user_id"`
	Username      pgtype.Text     `json:"username"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Sessions      []Sessions      `json:"sessions"`
	Tokens        []Tokens        `json:"tokens"`
	Organizations []Organizations `json:"organizations"`
}

// FindUserByID implements Querier.FindUserByID.
func (q *DBQuerier) FindUserByID(ctx context.Context, userID pgtype.Text) (FindUserByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserByID")
	row := q.conn.QueryRow(ctx, findUserByIDSQL, userID)
	var item FindUserByIDRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("query FindUserByID: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserByID row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserByID row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserByID row: %w", err)
	}
	return item, nil
}

// FindUserByIDBatch implements Querier.FindUserByIDBatch.
func (q *DBQuerier) FindUserByIDBatch(batch genericBatch, userID pgtype.Text) {
	batch.Queue(findUserByIDSQL, userID)
}

// FindUserByIDScan implements Querier.FindUserByIDScan.
func (q *DBQuerier) FindUserByIDScan(results pgx.BatchResults) (FindUserByIDRow, error) {
	row := results.QueryRow()
	var item FindUserByIDRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("scan FindUserByIDBatch row: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserByID row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserByID row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserByID row: %w", err)
	}
	return item, nil
}

const findUserByUsernameSQL = `SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE u.username = $1
GROUP BY u.user_id
;`

type FindUserByUsernameRow struct {
	UserID        pgtype.Text     `json:"user_id"`
	Username      pgtype.Text     `json:"username"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Sessions      []Sessions      `json:"sessions"`
	Tokens        []Tokens        `json:"tokens"`
	Organizations []Organizations `json:"organizations"`
}

// FindUserByUsername implements Querier.FindUserByUsername.
func (q *DBQuerier) FindUserByUsername(ctx context.Context, username pgtype.Text) (FindUserByUsernameRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserByUsername")
	row := q.conn.QueryRow(ctx, findUserByUsernameSQL, username)
	var item FindUserByUsernameRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("query FindUserByUsername: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserByUsername row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserByUsername row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserByUsername row: %w", err)
	}
	return item, nil
}

// FindUserByUsernameBatch implements Querier.FindUserByUsernameBatch.
func (q *DBQuerier) FindUserByUsernameBatch(batch genericBatch, username pgtype.Text) {
	batch.Queue(findUserByUsernameSQL, username)
}

// FindUserByUsernameScan implements Querier.FindUserByUsernameScan.
func (q *DBQuerier) FindUserByUsernameScan(results pgx.BatchResults) (FindUserByUsernameRow, error) {
	row := results.QueryRow()
	var item FindUserByUsernameRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("scan FindUserByUsernameBatch row: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserByUsername row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserByUsername row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserByUsername row: %w", err)
	}
	return item, nil
}

const findUserBySessionTokenSQL = `SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE s.token = $1
GROUP BY u.user_id
;`

type FindUserBySessionTokenRow struct {
	UserID        pgtype.Text     `json:"user_id"`
	Username      pgtype.Text     `json:"username"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Sessions      []Sessions      `json:"sessions"`
	Tokens        []Tokens        `json:"tokens"`
	Organizations []Organizations `json:"organizations"`
}

// FindUserBySessionToken implements Querier.FindUserBySessionToken.
func (q *DBQuerier) FindUserBySessionToken(ctx context.Context, token pgtype.Text) (FindUserBySessionTokenRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserBySessionToken")
	row := q.conn.QueryRow(ctx, findUserBySessionTokenSQL, token)
	var item FindUserBySessionTokenRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("query FindUserBySessionToken: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserBySessionToken row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserBySessionToken row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserBySessionToken row: %w", err)
	}
	return item, nil
}

// FindUserBySessionTokenBatch implements Querier.FindUserBySessionTokenBatch.
func (q *DBQuerier) FindUserBySessionTokenBatch(batch genericBatch, token pgtype.Text) {
	batch.Queue(findUserBySessionTokenSQL, token)
}

// FindUserBySessionTokenScan implements Querier.FindUserBySessionTokenScan.
func (q *DBQuerier) FindUserBySessionTokenScan(results pgx.BatchResults) (FindUserBySessionTokenRow, error) {
	row := results.QueryRow()
	var item FindUserBySessionTokenRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("scan FindUserBySessionTokenBatch row: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserBySessionToken row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserBySessionToken row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserBySessionToken row: %w", err)
	}
	return item, nil
}

const findUserByAuthenticationTokenSQL = `SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE t.token = $1
GROUP BY u.user_id
;`

type FindUserByAuthenticationTokenRow struct {
	UserID        pgtype.Text     `json:"user_id"`
	Username      pgtype.Text     `json:"username"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Sessions      []Sessions      `json:"sessions"`
	Tokens        []Tokens        `json:"tokens"`
	Organizations []Organizations `json:"organizations"`
}

// FindUserByAuthenticationToken implements Querier.FindUserByAuthenticationToken.
func (q *DBQuerier) FindUserByAuthenticationToken(ctx context.Context, token pgtype.Text) (FindUserByAuthenticationTokenRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserByAuthenticationToken")
	row := q.conn.QueryRow(ctx, findUserByAuthenticationTokenSQL, token)
	var item FindUserByAuthenticationTokenRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("query FindUserByAuthenticationToken: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationToken row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationToken row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationToken row: %w", err)
	}
	return item, nil
}

// FindUserByAuthenticationTokenBatch implements Querier.FindUserByAuthenticationTokenBatch.
func (q *DBQuerier) FindUserByAuthenticationTokenBatch(batch genericBatch, token pgtype.Text) {
	batch.Queue(findUserByAuthenticationTokenSQL, token)
}

// FindUserByAuthenticationTokenScan implements Querier.FindUserByAuthenticationTokenScan.
func (q *DBQuerier) FindUserByAuthenticationTokenScan(results pgx.BatchResults) (FindUserByAuthenticationTokenRow, error) {
	row := results.QueryRow()
	var item FindUserByAuthenticationTokenRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("scan FindUserByAuthenticationTokenBatch row: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationToken row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationToken row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationToken row: %w", err)
	}
	return item, nil
}

const findUserByAuthenticationTokenIDSQL = `SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s USING(user_id)
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE t.token_id = $1
GROUP BY u.user_id
;`

type FindUserByAuthenticationTokenIDRow struct {
	UserID        pgtype.Text     `json:"user_id"`
	Username      pgtype.Text     `json:"username"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Sessions      []Sessions      `json:"sessions"`
	Tokens        []Tokens        `json:"tokens"`
	Organizations []Organizations `json:"organizations"`
}

// FindUserByAuthenticationTokenID implements Querier.FindUserByAuthenticationTokenID.
func (q *DBQuerier) FindUserByAuthenticationTokenID(ctx context.Context, tokenID pgtype.Text) (FindUserByAuthenticationTokenIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindUserByAuthenticationTokenID")
	row := q.conn.QueryRow(ctx, findUserByAuthenticationTokenIDSQL, tokenID)
	var item FindUserByAuthenticationTokenIDRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("query FindUserByAuthenticationTokenID: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationTokenID row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationTokenID row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationTokenID row: %w", err)
	}
	return item, nil
}

// FindUserByAuthenticationTokenIDBatch implements Querier.FindUserByAuthenticationTokenIDBatch.
func (q *DBQuerier) FindUserByAuthenticationTokenIDBatch(batch genericBatch, tokenID pgtype.Text) {
	batch.Queue(findUserByAuthenticationTokenIDSQL, tokenID)
}

// FindUserByAuthenticationTokenIDScan implements Querier.FindUserByAuthenticationTokenIDScan.
func (q *DBQuerier) FindUserByAuthenticationTokenIDScan(results pgx.BatchResults) (FindUserByAuthenticationTokenIDRow, error) {
	row := results.QueryRow()
	var item FindUserByAuthenticationTokenIDRow
	sessionsArray := q.types.newSessionsArray()
	tokensArray := q.types.newTokensArray()
	organizationsArray := q.types.newOrganizationsArray()
	if err := row.Scan(&item.UserID, &item.Username, &item.CreatedAt, &item.UpdatedAt, sessionsArray, tokensArray, organizationsArray); err != nil {
		return item, fmt.Errorf("scan FindUserByAuthenticationTokenIDBatch row: %w", err)
	}
	if err := sessionsArray.AssignTo(&item.Sessions); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationTokenID row: %w", err)
	}
	if err := tokensArray.AssignTo(&item.Tokens); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationTokenID row: %w", err)
	}
	if err := organizationsArray.AssignTo(&item.Organizations); err != nil {
		return item, fmt.Errorf("assign FindUserByAuthenticationTokenID row: %w", err)
	}
	return item, nil
}

const deleteUserByIDSQL = `DELETE
FROM users
WHERE user_id = $1
RETURNING user_id
;`

// DeleteUserByID implements Querier.DeleteUserByID.
func (q *DBQuerier) DeleteUserByID(ctx context.Context, userID pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteUserByID")
	row := q.conn.QueryRow(ctx, deleteUserByIDSQL, userID)
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query DeleteUserByID: %w", err)
	}
	return item, nil
}

// DeleteUserByIDBatch implements Querier.DeleteUserByIDBatch.
func (q *DBQuerier) DeleteUserByIDBatch(batch genericBatch, userID pgtype.Text) {
	batch.Queue(deleteUserByIDSQL, userID)
}

// DeleteUserByIDScan implements Querier.DeleteUserByIDScan.
func (q *DBQuerier) DeleteUserByIDScan(results pgx.BatchResults) (pgtype.Text, error) {
	row := results.QueryRow()
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan DeleteUserByIDBatch row: %w", err)
	}
	return item, nil
}

const deleteUserByUsernameSQL = `DELETE
FROM users
WHERE username = $1
RETURNING user_id
;`

// DeleteUserByUsername implements Querier.DeleteUserByUsername.
func (q *DBQuerier) DeleteUserByUsername(ctx context.Context, username pgtype.Text) (pgtype.Text, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteUserByUsername")
	row := q.conn.QueryRow(ctx, deleteUserByUsernameSQL, username)
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query DeleteUserByUsername: %w", err)
	}
	return item, nil
}

// DeleteUserByUsernameBatch implements Querier.DeleteUserByUsernameBatch.
func (q *DBQuerier) DeleteUserByUsernameBatch(batch genericBatch, username pgtype.Text) {
	batch.Queue(deleteUserByUsernameSQL, username)
}

// DeleteUserByUsernameScan implements Querier.DeleteUserByUsernameScan.
func (q *DBQuerier) DeleteUserByUsernameScan(results pgx.BatchResults) (pgtype.Text, error) {
	row := results.QueryRow()
	var item pgtype.Text
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan DeleteUserByUsernameBatch row: %w", err)
	}
	return item, nil
}