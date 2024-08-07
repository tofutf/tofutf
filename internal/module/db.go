package module

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal/connections"
	"github.com/tofutf/tofutf/internal/semver"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type (
	// pgdb is the registry database on postgres
	pgdb struct {
		*sql.Pool // provides access to generated SQL queries
	}

	// moduleRow is a row from a database query for modules.
	moduleRow struct {
		ModuleID         pgtype.Text             `json:"module_id"`
		CreatedAt        pgtype.Timestamptz      `json:"created_at"`
		UpdatedAt        pgtype.Timestamptz      `json:"updated_at"`
		Name             pgtype.Text             `json:"name"`
		Provider         pgtype.Text             `json:"provider"`
		Status           pgtype.Text             `json:"status"`
		OrganizationName pgtype.Text             `json:"organization_name"`
		ModuleConnection *pggen.RepoConnections  `json:"module_connection"`
		Versions         []*pggen.ModuleVersions `json:"versions"`
	}
)

func (db *pgdb) createModule(ctx context.Context, mod *Module) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertModule(ctx, pggen.InsertModuleParams{
			ID:               sql.String(mod.ID),
			CreatedAt:        sql.Timestamptz(mod.CreatedAt),
			UpdatedAt:        sql.Timestamptz(mod.UpdatedAt),
			Name:             sql.String(mod.Name),
			Provider:         sql.String(mod.Provider),
			Status:           sql.String(string(mod.Status)),
			OrganizationName: sql.String(mod.Organization),
		})
		return sql.Error(err)
	})
}

func (db *pgdb) updateModuleStatus(ctx context.Context, moduleID string, status ModuleStatus) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.UpdateModuleStatusByID(ctx, sql.String(string(status)), sql.String(moduleID))
		if err != nil {
			return sql.Error(err)
		}
		return nil
	})
}

func (db *pgdb) listModules(ctx context.Context, opts ListModulesOptions) ([]*Module, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]*Module, error) {
		rows, err := q.ListModulesByOrganization(ctx, sql.String(opts.Organization))
		if err != nil {
			return nil, err
		}

		modules := make([]*Module, len(rows))
		for i, r := range rows {
			modules[i] = moduleRow(r).toModule()
		}
		return modules, nil
	})
}

func (db *pgdb) getModule(ctx context.Context, opts GetModuleOptions) (*Module, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*Module, error) {
		row, err := q.FindModuleByName(ctx, pggen.FindModuleByNameParams{
			Name:             sql.String(opts.Name),
			Provider:         sql.String(opts.Provider),
			OrganizationName: sql.String(opts.Organization),
		})
		if err != nil {
			return nil, sql.Error(err)
		}

		return moduleRow(row).toModule(), nil
	})
}

func (db *pgdb) getModuleByID(ctx context.Context, id string) (*Module, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*Module, error) {
		row, err := q.FindModuleByID(ctx, sql.String(id))
		if err != nil {
			return nil, sql.Error(err)
		}

		return moduleRow(row).toModule(), nil
	})
}

func (db *pgdb) getModuleByConnection(ctx context.Context, vcsProviderID, repoPath string) (*Module, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*Module, error) {
		row, err := q.FindModuleByConnection(ctx, sql.String(vcsProviderID), sql.String(repoPath))
		if err != nil {
			return nil, sql.Error(err)
		}

		return moduleRow(row).toModule(), nil
	})
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteModuleByID(ctx, sql.String(id))
		return sql.Error(err)
	})
}

func (db *pgdb) createModuleVersion(ctx context.Context, version *ModuleVersion) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertModuleVersion(ctx, pggen.InsertModuleVersionParams{
			ModuleVersionID: sql.String(version.ID),
			Version:         sql.String(version.Version),
			CreatedAt:       sql.Timestamptz(version.CreatedAt),
			UpdatedAt:       sql.Timestamptz(version.UpdatedAt),
			ModuleID:        sql.String(version.ModuleID),
			Status:          sql.String(string(version.Status)),
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *pgdb) updateModuleVersionStatus(ctx context.Context, opts UpdateModuleVersionStatusOptions) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.UpdateModuleVersionStatusByID(ctx, pggen.UpdateModuleVersionStatusByIDParams{
			ModuleVersionID: sql.String(opts.ID),
			Status:          sql.String(string(opts.Status)),
			StatusError:     sql.String(opts.Error),
		})
		return sql.Error(err)
	})
}

func (db *pgdb) getModuleByVersionID(ctx context.Context, versionID string) (*Module, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*Module, error) {
		row, err := q.FindModuleByModuleVersionID(ctx, sql.String(versionID))
		if err != nil {
			return nil, sql.Error(err)
		}

		return moduleRow(row).toModule(), nil
	})
}

func (db *pgdb) deleteModuleVersion(ctx context.Context, versionID string) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteModuleVersionByID(ctx, sql.String(versionID))
		return sql.Error(err)
	})
}

func (db *pgdb) saveTarball(ctx context.Context, versionID string, tarball []byte) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertModuleTarball(ctx, tarball, sql.String(versionID))
		return sql.Error(err)
	})
}

func (db *pgdb) getTarball(ctx context.Context, versionID string) ([]byte, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]byte, error) {
		tarball, err := q.FindModuleTarball(ctx, sql.String(versionID))
		if err != nil {
			return nil, sql.Error(err)
		}

		return tarball, nil
	})
}

// toModule converts a database row into a module
func (row moduleRow) toModule() *Module {
	module := &Module{
		ID:           row.ModuleID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		UpdatedAt:    row.UpdatedAt.Time.UTC(),
		Name:         row.Name.String,
		Provider:     row.Provider.String,
		Status:       ModuleStatus(row.Status.String),
		Organization: row.OrganizationName.String,
	}
	if row.ModuleConnection != nil {
		module.Connection = &connections.Connection{
			VCSProviderID: row.ModuleConnection.VCSProviderID.String,
			Repo:          row.ModuleConnection.RepoPath.String,
		}
	}
	// versions are always maintained in descending order
	sort.Sort(byVersion(row.Versions))
	for i := len(row.Versions) - 1; i >= 0; i-- {
		module.Versions = append(module.Versions, ModuleVersion{
			ID:          row.Versions[i].ModuleVersionID.String,
			Version:     row.Versions[i].Version.String,
			CreatedAt:   row.Versions[i].CreatedAt.Time.UTC(),
			UpdatedAt:   row.Versions[i].UpdatedAt.Time.UTC(),
			ModuleID:    row.Versions[i].ModuleID.String,
			Status:      ModuleVersionStatus(row.Versions[i].Status.String),
			StatusError: row.Versions[i].StatusError.String,
		})
	}
	return module
}

type byVersion []*pggen.ModuleVersions

func (v byVersion) Len() int      { return len(v) }
func (v byVersion) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v byVersion) Less(i, j int) bool {
	return semver.Compare(v[i].Version.String, v[j].Version.String) < 0
}
