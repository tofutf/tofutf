package configversion

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/resource"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type pgdb struct {
	*sql.Pool // provides access to generated SQL queries
}

func (db *pgdb) CreateConfigurationVersion(ctx context.Context, cv *ConfigurationVersion) error {
	return db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertConfigurationVersion(ctx, pggen.InsertConfigurationVersionParams{
			ID:            sql.String(cv.ID),
			CreatedAt:     sql.Timestamptz(cv.CreatedAt),
			AutoQueueRuns: sql.Bool(cv.AutoQueueRuns),
			Source:        sql.String(string(cv.Source)),
			Speculative:   sql.Bool(cv.Speculative),
			Status:        sql.String(string(cv.Status)),
			WorkspaceID:   sql.String(cv.WorkspaceID),
		})
		if err != nil {
			return fmt.Errorf("failed to insert configuration version: %w", err)
		}

		if cv.IngressAttributes != nil {
			ia := cv.IngressAttributes
			_, err := q.InsertIngressAttributes(ctx, pggen.InsertIngressAttributesParams{
				Branch:                 sql.String(ia.Branch),
				CommitSHA:              sql.String(ia.CommitSHA),
				CommitURL:              sql.String(ia.CommitURL),
				PullRequestNumber:      sql.Int4(ia.PullRequestNumber),
				PullRequestURL:         sql.String(ia.PullRequestURL),
				PullRequestTitle:       sql.String(ia.PullRequestTitle),
				SenderUsername:         sql.String(ia.SenderUsername),
				SenderAvatarURL:        sql.String(ia.SenderAvatarURL),
				SenderHTMLURL:          sql.String(ia.SenderHTMLURL),
				Tag:                    sql.String(ia.Tag),
				Identifier:             sql.String(ia.Repo),
				IsPullRequest:          sql.Bool(ia.IsPullRequest),
				OnDefaultBranch:        sql.Bool(ia.OnDefaultBranch),
				ConfigurationVersionID: sql.String(cv.ID),
			})
			if err != nil {
				return fmt.Errorf("failed to insert ingress attribute: %w", err)
			}
		}

		// Insert timestamp for current status
		if err := db.insertCVStatusTimestamp(ctx, cv); err != nil {
			return fmt.Errorf("inserting configuration version status timestamp: %w", err)
		}
		return nil
	})
}

func (db *pgdb) UploadConfigurationVersion(ctx context.Context, id string, fn func(*ConfigurationVersion, ConfigUploader) error) error {
	return db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		// select ...for update
		result, err := q.FindConfigurationVersionByIDForUpdate(ctx, sql.String(id))
		if err != nil {
			return fmt.Errorf("failed to find configuration version to modify: %w", err)
		}
		cv := pgRow(result).toConfigVersion()

		if err := fn(cv, newConfigUploader(q, cv.ID)); err != nil {
			return fmt.Errorf("failed to mutate configuration version: %w", err)
		}

		return nil
	})
}

func (db *pgdb) ListConfigurationVersions(ctx context.Context, workspaceID string, opts ListOptions) (*resource.Page[*ConfigurationVersion], error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*resource.Page[*ConfigurationVersion], error) {
		rows, err := q.FindConfigurationVersionsByWorkspaceID(ctx, pggen.FindConfigurationVersionsByWorkspaceIDParams{
			WorkspaceID: sql.String(workspaceID),
			Limit:       opts.GetLimit(),
			Offset:      opts.GetOffset(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list configuration versions in workspace: %w", err)
		}

		count, err := q.CountConfigurationVersionsByWorkspaceID(ctx, sql.String(workspaceID))
		if err != nil {
			return nil, fmt.Errorf("failed to count configuration versions in workspace: %w", err)
		}

		items := make([]*ConfigurationVersion, len(rows))
		for i, r := range rows {
			items[i] = pgRow(r).toConfigVersion()
		}

		return resource.NewPage(items, opts.PageOptions, internal.Int64(count.Int64)), nil
	})
}

func (db *pgdb) GetConfigurationVersion(ctx context.Context, opts ConfigurationVersionGetOptions) (*ConfigurationVersion, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*ConfigurationVersion, error) {
		if opts.ID != nil {
			result, err := q.FindConfigurationVersionByID(ctx, sql.String(*opts.ID))
			if err != nil {
				return nil, fmt.Errorf("failed to find configuration by id: %w", sql.Error(err))
			}

			return pgRow(result).toConfigVersion(), nil
		} else if opts.WorkspaceID != nil {
			result, err := q.FindConfigurationVersionLatestByWorkspaceID(ctx, sql.String(*opts.WorkspaceID))
			if err != nil {
				return nil, fmt.Errorf("failed to find configuration by workspace id: %w", sql.Error(err))
			}

			return pgRow(result).toConfigVersion(), nil
		}

		return nil, fmt.Errorf("no configuration version spec provided")
	})
}

func (db *pgdb) GetConfig(ctx context.Context, id string) ([]byte, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]byte, error) {
		cfg, err := q.DownloadConfigurationVersion(ctx, sql.String(id))
		if err != nil {
			return nil, fmt.Errorf("failed to download configuration version tarball: %w", sql.Error(err))
		}

		return cfg, nil
	})
}

func (db *pgdb) DeleteConfigurationVersion(ctx context.Context, id string) error {
	return db.Pool.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteConfigurationVersionByID(ctx, sql.String(id))
		if err != nil {
			return fmt.Errorf("failed to delete configuration version by id: %w", sql.Error(err))
		}

		return nil
	})
}

func (db *pgdb) insertCVStatusTimestamp(ctx context.Context, cv *ConfigurationVersion) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		sts, err := cv.StatusTimestamp(cv.Status)
		if err != nil {
			return fmt.Errorf("failed to get status timestamp: %w", err)
		}

		_, err = q.InsertConfigurationVersionStatusTimestamp(ctx, pggen.InsertConfigurationVersionStatusTimestampParams{
			ID:        sql.String(cv.ID),
			Status:    sql.String(string(cv.Status)),
			Timestamp: sql.Timestamptz(sts),
		})
		if err != nil {
			return fmt.Errorf("failed to insert configuration version status timestamp: %w", err)
		}

		return nil
	})
}

// pgRow represents the result of a database query for a configuration version.
type pgRow struct {
	ConfigurationVersionID               pgtype.Text                                   `json:"configuration_version_id"`
	CreatedAt                            pgtype.Timestamptz                            `json:"created_at"`
	AutoQueueRuns                        pgtype.Bool                                   `json:"auto_queue_runs"`
	Source                               pgtype.Text                                   `json:"source"`
	Speculative                          pgtype.Bool                                   `json:"speculative"`
	Status                               pgtype.Text                                   `json:"status"`
	WorkspaceID                          pgtype.Text                                   `json:"workspace_id"`
	ConfigurationVersionStatusTimestamps []*pggen.ConfigurationVersionStatusTimestamps `json:"configuration_version_status_timestamps"`
	IngressAttributes                    *pggen.IngressAttributes                      `json:"ingress_attributes"`
}

func (result pgRow) toConfigVersion() *ConfigurationVersion {
	cv := ConfigurationVersion{
		ID:               result.ConfigurationVersionID.String,
		CreatedAt:        result.CreatedAt.Time.UTC(),
		AutoQueueRuns:    result.AutoQueueRuns.Bool,
		Speculative:      result.Speculative.Bool,
		Source:           Source(result.Source.String),
		Status:           ConfigurationStatus(result.Status.String),
		StatusTimestamps: unmarshalStatusTimestampRows(result.ConfigurationVersionStatusTimestamps),
		WorkspaceID:      result.WorkspaceID.String,
	}
	if result.IngressAttributes != nil {
		cv.IngressAttributes = NewIngressFromRow(result.IngressAttributes)
	}
	return &cv
}

func NewIngressFromRow(row *pggen.IngressAttributes) *IngressAttributes {
	return &IngressAttributes{
		Branch:            row.Branch.String,
		CommitSHA:         row.CommitSHA.String,
		CommitURL:         row.CommitURL.String,
		Repo:              row.Identifier.String,
		IsPullRequest:     row.IsPullRequest.Bool,
		PullRequestNumber: int(row.PullRequestNumber.Int32),
		PullRequestURL:    row.PullRequestURL.String,
		PullRequestTitle:  row.PullRequestTitle.String,
		SenderUsername:    row.SenderUsername.String,
		SenderAvatarURL:   row.SenderAvatarURL.String,
		SenderHTMLURL:     row.SenderHTMLURL.String,
		Tag:               row.Tag.String,
		OnDefaultBranch:   row.IsPullRequest.Bool,
	}
}

func unmarshalStatusTimestampRows(rows []*pggen.ConfigurationVersionStatusTimestamps) (timestamps []ConfigurationVersionStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, ConfigurationVersionStatusTimestamp{
			Status:    ConfigurationStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}
