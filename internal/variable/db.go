package variable

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type (
	// pgdb is a database of variables on postgres
	pgdb struct {
		*sql.Pool // provides access to generated SQL queries
	}

	variableRow struct {
		VariableID  pgtype.Text `json:"variable_id"`
		Key         pgtype.Text `json:"key"`
		Value       pgtype.Text `json:"value"`
		Description pgtype.Text `json:"description"`
		Category    pgtype.Text `json:"category"`
		Sensitive   pgtype.Bool `json:"sensitive"`
		HCL         pgtype.Bool `json:"hcl"`
		VersionID   pgtype.Text `json:"version_id"`
	}

	variableSetRow struct {
		VariableSetID    pgtype.Text        `json:"variable_set_id"`
		Global           pgtype.Bool        `json:"global"`
		Name             pgtype.Text        `json:"name"`
		Description      pgtype.Text        `json:"description"`
		OrganizationName pgtype.Text        `json:"organization_name"`
		Variables        []*pggen.Variables `json:"variables"`
		WorkspaceIds     []string           `json:"workspace_ids"`
	}
)

func (row variableRow) convert() *Variable {
	return &Variable{
		ID:          row.VariableID.String,
		Key:         row.Key.String,
		Value:       row.Value.String,
		Description: row.Description.String,
		Category:    VariableCategory(row.Category.String),
		Sensitive:   row.Sensitive.Bool,
		HCL:         row.HCL.Bool,
		VersionID:   row.VersionID.String,
	}
}

func (row variableSetRow) convert() *VariableSet {
	set := &VariableSet{
		ID:           row.VariableSetID.String,
		Global:       row.Global.Bool,
		Description:  row.Description.String,
		Name:         row.Name.String,
		Organization: row.OrganizationName.String,
	}
	set.Variables = make([]*Variable, len(row.Variables))
	for i, v := range row.Variables {
		set.Variables[i] = variableRow(*v).convert()
	}
	set.Workspaces = row.WorkspaceIds
	return set
}

func (pdb *pgdb) createWorkspaceVariable(ctx context.Context, workspaceID string, v *Variable) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		_, err := q.InsertWorkspaceVariable(ctx, sql.String(v.ID), sql.String(workspaceID))
		return err
	})
	return sql.Error(err)
}

func (pdb *pgdb) listWorkspaceVariables(ctx context.Context, workspaceID string) ([]*Variable, error) {
	return sql.Query(ctx, pdb.Pool, func(ctx context.Context, q pggen.Querier) ([]*Variable, error) {
		rows, err := q.FindWorkspaceVariablesByWorkspaceID(ctx, sql.String(workspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}

		variables := make([]*Variable, len(rows))
		for i, row := range rows {
			variables[i] = variableRow(row).convert()
		}

		return variables, nil
	})
}

func (pdb *pgdb) getWorkspaceVariable(ctx context.Context, variableID string) (*WorkspaceVariable, error) {
	return sql.Query(ctx, pdb.Pool, func(ctx context.Context, q pggen.Querier) (*WorkspaceVariable, error) {
		row, err := q.FindWorkspaceVariableByVariableID(ctx, sql.String(variableID))
		if err != nil {
			return nil, sql.Error(err)
		}

		return &WorkspaceVariable{
			WorkspaceID: row.WorkspaceID.String,
			Variable:    variableRow(*row.Variable).convert(),
		}, nil
	})
}

func (pdb *pgdb) deleteWorkspaceVariable(ctx context.Context, variableID string) (*WorkspaceVariable, error) {
	return sql.Query(ctx, pdb.Pool, func(ctx context.Context, q pggen.Querier) (*WorkspaceVariable, error) {
		row, err := q.DeleteWorkspaceVariableByID(ctx, sql.String(variableID))
		if err != nil {
			return nil, sql.Error(err)
		}

		return &WorkspaceVariable{
			WorkspaceID: row.WorkspaceID.String,
			Variable:    variableRow(*row.Variable).convert(),
		}, nil
	})
}

func (pdb *pgdb) createVariableSet(ctx context.Context, set *VariableSet) error {
	return pdb.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertVariableSet(ctx, pggen.InsertVariableSetParams{
			VariableSetID:    sql.String(set.ID),
			Name:             sql.String(set.Name),
			Description:      sql.String(set.Description),
			Global:           sql.Bool(set.Global),
			OrganizationName: sql.String(set.Organization),
		})
		return sql.Error(err)
	})
}

func (pdb *pgdb) updateVariableSet(ctx context.Context, set *VariableSet) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.UpdateVariableSetByID(ctx, pggen.UpdateVariableSetByIDParams{
			Name:          sql.String(set.Name),
			Description:   sql.String(set.Description),
			Global:        sql.Bool(set.Global),
			VariableSetID: sql.String(set.ID),
		})
		if err != nil {
			return err
		}

		// lazily delete all variable set workspaces, and then add them again,
		// regardless of whether there are any changes
		return pdb.Lock(ctx, "variable_set_workspaces", func(ctx context.Context, q pggen.Querier) error {
			if err := pdb.deleteAllVariableSetWorkspaces(ctx, set.ID); err != nil {
				return err
			}
			if err := pdb.createVariableSetWorkspaces(ctx, set.ID, set.Workspaces); err != nil {
				return err
			}
			return nil
		})
	})
	return sql.Error(err)
}

func (pdb *pgdb) getVariableSet(ctx context.Context, setID string) (*VariableSet, error) {
	return sql.Query(ctx, pdb.Pool, func(ctx context.Context, q pggen.Querier) (*VariableSet, error) {
		row, err := q.FindVariableSetBySetID(ctx, sql.String(setID))
		if err != nil {
			return nil, sql.Error(err)
		}

		return variableSetRow(row).convert(), nil
	})
}

func (pdb *pgdb) getVariableSetByVariableID(ctx context.Context, variableID string) (*VariableSet, error) {
	return sql.Query(ctx, pdb.Pool, func(ctx context.Context, q pggen.Querier) (*VariableSet, error) {
		row, err := q.FindVariableSetByVariableID(ctx, sql.String(variableID))
		if err != nil {
			return nil, sql.Error(err)
		}

		return variableSetRow(row).convert(), nil
	})
}

func (pdb *pgdb) listVariableSets(ctx context.Context, organization string) ([]*VariableSet, error) {
	return sql.Query(ctx, pdb.Pool, func(ctx context.Context, q pggen.Querier) ([]*VariableSet, error) {
		rows, err := q.FindVariableSetsByOrganization(ctx, sql.String(organization))
		if err != nil {
			return nil, sql.Error(err)
		}

		sets := make([]*VariableSet, len(rows))
		for i, row := range rows {
			sets[i] = variableSetRow(row).convert()
		}

		return sets, nil
	})
}

func (pdb *pgdb) listVariableSetsByWorkspace(ctx context.Context, workspaceID string) ([]*VariableSet, error) {
	return sql.Query(ctx, pdb.Pool, func(ctx context.Context, q pggen.Querier) ([]*VariableSet, error) {
		rows, err := q.FindVariableSetsByWorkspace(ctx, sql.String(workspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}

		sets := make([]*VariableSet, len(rows))
		for i, row := range rows {
			sets[i] = variableSetRow(row).convert()
		}

		return sets, nil
	})
}

func (pdb *pgdb) deleteVariableSet(ctx context.Context, setID string) error {
	return pdb.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteVariableSetByID(ctx, sql.String(setID))
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}

func (pdb *pgdb) addVariableToSet(ctx context.Context, setID string, v *Variable) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		_, err := q.InsertVariableSetVariable(ctx, sql.String(setID), sql.String(v.ID))
		return err
	})
	return sql.Error(err)
}

func (pdb *pgdb) createVariableSetWorkspaces(ctx context.Context, setID string, workspaceIDs []string) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		for _, wid := range workspaceIDs {
			_, err := q.InsertVariableSetWorkspace(ctx, sql.String(setID), sql.String(wid))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (pdb *pgdb) deleteAllVariableSetWorkspaces(ctx context.Context, setID string) error {
	return pdb.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteVariableSetWorkspaces(ctx, sql.String(setID))
		return sql.Error(err)
	})
}

func (pdb *pgdb) deleteVariableSetWorkspaces(ctx context.Context, setID string, workspaceIDs []string) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		for _, wid := range workspaceIDs {
			_, err := q.DeleteVariableSetWorkspace(ctx, sql.String(setID), sql.String(wid))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (pdb *pgdb) createVariable(ctx context.Context, v *Variable) error {
	return pdb.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertVariable(ctx, pggen.InsertVariableParams{
			VariableID:  sql.String(v.ID),
			Key:         sql.String(v.Key),
			Value:       sql.String(v.Value),
			Description: sql.String(v.Description),
			Category:    sql.String(string(v.Category)),
			Sensitive:   sql.Bool(v.Sensitive),
			VersionID:   sql.String(v.VersionID),
			HCL:         sql.Bool(v.HCL),
		})
		return sql.Error(err)
	})
}

func (pdb *pgdb) updateVariable(ctx context.Context, v *Variable) error {
	return pdb.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.UpdateVariableByID(ctx, pggen.UpdateVariableByIDParams{
			VariableID:  sql.String(v.ID),
			Key:         sql.String(v.Key),
			Value:       sql.String(v.Value),
			Description: sql.String(v.Description),
			Category:    sql.String(string(v.Category)),
			Sensitive:   sql.Bool(v.Sensitive),
			VersionID:   sql.String(v.VersionID),
			HCL:         sql.Bool(v.HCL),
		})
		return sql.Error(err)
	})
}

func (pdb *pgdb) deleteVariable(ctx context.Context, variableID string) error {
	return pdb.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteVariableByID(ctx, sql.String(variableID))
		return sql.Error(err)
	})
}
