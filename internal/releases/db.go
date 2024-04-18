package releases

import (
	"context"
	"time"

	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type db struct {
	*sql.Pool
}

func (db *db) updateLatestVersion(ctx context.Context, v string) error {
	return db.Lock(ctx, "latest_terraform_version", func(ctx context.Context, q pggen.Querier) error {
		rows, err := q.FindLatestTerraformVersion(ctx)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			_, err = q.InsertLatestTerraformVersion(ctx, sql.String(v))
			if err != nil {
				return err
			}
		} else {
			_, err = q.UpdateLatestTerraformVersion(ctx, sql.String(v))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *db) getLatest(ctx context.Context) (string, time.Time, error) {
	type latestRelease struct {
		Version    string
		Checkpoint time.Time
	}

	latest, err := sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (latestRelease, error) {
		rows, err := q.FindLatestTerraformVersion(ctx)
		if err != nil {
			return latestRelease{}, err
		}

		if len(rows) == 0 {
			return latestRelease{}, internal.ErrResourceNotFound
		}

		return latestRelease{rows[0].Version.String, rows[0].Checkpoint.Time}, nil
	})
	if err != nil {
		return "", time.Time{}, err
	}

	return latest.Version, latest.Checkpoint, nil
}
