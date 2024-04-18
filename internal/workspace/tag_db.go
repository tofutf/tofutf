package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/resource"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

type (
	// pgresult represents the result of a database query for a tag.
	tagresult struct {
		TagID            pgtype.Text `json:"tag_id"`
		Name             pgtype.Text `json:"name"`
		OrganizationName pgtype.Text `json:"organization_name"`
		InstanceCount    pgtype.Int8 `json:"instance_count"`
	}
)

// toTag converts a database result into a tag
func (r tagresult) toTag() *Tag {
	return &Tag{
		ID:            r.TagID.String,
		Name:          r.Name.String,
		Organization:  r.OrganizationName.String,
		InstanceCount: int(r.InstanceCount.Int64),
	}
}

func (db *pgdb) listTags(ctx context.Context, organization string, opts ListTagsOptions) (*resource.Page[*Tag], error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*resource.Page[*Tag], error) {
		rows, err := q.FindTags(ctx, pggen.FindTagsParams{
			OrganizationName: sql.String(organization),
			Limit:            opts.GetLimit(),
			Offset:           opts.GetOffset(),
		})
		if err != nil {
			return nil, sql.Error(err)
		}

		count, err := q.CountTags(ctx, sql.String(organization))
		if err != nil {
			return nil, sql.Error(err)
		}

		items := make([]*Tag, len(rows))
		for i, r := range rows {
			items[i] = tagresult(r).toTag()
		}

		return resource.NewPage(items, opts.PageOptions, internal.Int64(count.Int64)), nil
	})
}

func (db *pgdb) deleteTags(ctx context.Context, organization string, tagIDs []string) error {
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		for _, tid := range tagIDs {
			_, err := q.DeleteTag(ctx, sql.String(tid), sql.String(organization))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (db *pgdb) addTag(ctx context.Context, organization, name, id string) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertTag(ctx, pggen.InsertTagParams{
			TagID:            sql.String(id),
			Name:             sql.String(name),
			OrganizationName: sql.String(organization),
		})
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}

func (db *pgdb) findTagByName(ctx context.Context, organization, name string) (*Tag, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*Tag, error) {
		tag, err := q.FindTagByName(ctx, sql.String(name), sql.String(organization))
		if err != nil {
			return nil, sql.Error(err)
		}

		return tagresult(tag).toTag(), nil
	})
}

func (db *pgdb) findTagByID(ctx context.Context, organization, id string) (*Tag, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*Tag, error) {
		tag, err := q.FindTagByID(ctx, sql.String(id), sql.String(organization))
		if err != nil {
			return nil, sql.Error(err)
		}

		return tagresult(tag).toTag(), nil
	})
}

func (db *pgdb) tagWorkspace(ctx context.Context, workspaceID, tagID string) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertWorkspaceTag(ctx, sql.String(tagID), sql.String(workspaceID))
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}

func (db *pgdb) deleteWorkspaceTag(ctx context.Context, workspaceID, tagID string) error {
	return db.Query(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.DeleteWorkspaceTag(ctx, sql.String(workspaceID), sql.String(tagID))
		if err != nil {
			return sql.Error(err)
		}

		return nil
	})
}

func (db *pgdb) listWorkspaceTags(ctx context.Context, workspaceID string, opts ListWorkspaceTagsOptions) (*resource.Page[*Tag], error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (*resource.Page[*Tag], error) {
		rows, err := q.FindWorkspaceTags(ctx, pggen.FindWorkspaceTagsParams{
			WorkspaceID: sql.String(workspaceID),
			Limit:       opts.GetLimit(),
			Offset:      opts.GetOffset(),
		})
		if err != nil {
			return nil, sql.Error(err)
		}
		count, err := q.CountWorkspaceTags(ctx, sql.String(workspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}

		items := make([]*Tag, len(rows))
		for i, r := range rows {
			items[i] = tagresult(r).toTag()
		}

		return resource.NewPage(items, opts.PageOptions, internal.Int64(count.Int64)), nil
	})
}
