package workspace

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

func TestDB(t *testing.T) {
	pg := sql.NewTestContainer(t)
	ctx := context.Background()

	t.Run("create", func(t *testing.T) {
		pool := sql.TestContainerReset(t, pg)
		db := pgdb{pool}

		err := db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
			_, err := q.InsertOrganization(ctx, pggen.InsertOrganizationParams{
				ID:                         sql.String("organization"),
				Name:                       sql.String("organization"),
				CreatedAt:                  sql.Timestamptz(time.Now()),
				UpdatedAt:                  sql.Timestamptz(time.Now()),
				AllowForceDeleteWorkspaces: sql.Bool(true),
				CostEstimationEnabled:      sql.Bool(false),
			})
			return err
		})
		require.NoError(t, err)

		err = db.create(ctx, &Workspace{
			ID:                         "id",
			CreatedAt:                  time.Now(),
			UpdatedAt:                  time.Now(),
			AgentPoolID:                nil,
			AllowDestroyPlan:           true,
			AutoApply:                  false,
			Description:                "description",
			Environment:                "environment",
			ExecutionMode:              "remote",
			GlobalRemoteState:          false,
			MigrationEnvironment:       "migration-environment",
			Name:                       "workspace",
			QueueAllRuns:               true,
			SpeculativeEnabled:         true,
			StructuredRunOutputEnabled: true,
			SourceName:                 "src-name",
			SourceURL:                  "src-url",
			TerraformVersion:           "tf-version",
			WorkingDirectory:           "/",
			Organization:               "organization",
			LatestRun:                  nil,
			Tags:                       []string{"tag1", "tag2"},
			Connection:                 nil,
			TriggerPatterns:            nil,
			TriggerPrefixes:            nil,
		})
		require.NoError(t, err)

		ws, err := db.get(ctx, "id")
		require.NoError(t, err)

		_ = ws
	})
}
