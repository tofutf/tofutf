package logs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
)

// pgdb is a logs database on postgres
type pgdb struct {
	*sql.Pool // provides access to generated SQL queries
}

// put persists a chunk of logs to the DB and returns the chunk updated with a
// unique identifier

// put persists data to the DB and returns a unique identifier for the chunk
func (db *pgdb) put(ctx context.Context, opts internal.PutChunkOptions) (string, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (string, error) {
		if len(opts.Data) == 0 {
			return "", fmt.Errorf("refusing to persist empty chunk")
		}

		id, err := q.InsertLogChunk(ctx, pggen.InsertLogChunkParams{
			RunID:  sql.String(opts.RunID),
			Phase:  sql.String(string(opts.Phase)),
			Chunk:  opts.Data,
			Offset: sql.Int4(opts.Offset),
		})
		if err != nil {
			return "", sql.Error(err)
		}

		return strconv.Itoa(int(id.Int32)), nil
	})
}

func (db *pgdb) getChunk(ctx context.Context, chunkID string) (internal.Chunk, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) (internal.Chunk, error) {
		id, err := strconv.Atoi(chunkID)
		if err != nil {
			return internal.Chunk{}, err
		}

		chunk, err := q.FindLogChunkByID(ctx, sql.Int4(id))
		if err != nil {
			return internal.Chunk{}, sql.Error(err)
		}

		return internal.Chunk{
			ID:     chunkID,
			RunID:  chunk.RunID.String,
			Phase:  internal.PhaseType(chunk.Phase.String),
			Data:   chunk.Chunk,
			Offset: int(chunk.Offset.Int32),
		}, nil
	})
}

func (db *pgdb) getLogs(ctx context.Context, runID string, phase internal.PhaseType) ([]byte, error) {
	return sql.Query(ctx, db.Pool, func(ctx context.Context, q pggen.Querier) ([]byte, error) {
		data, err := q.FindLogs(ctx, sql.String(runID), sql.String(string(phase)))
		if err != nil {
			// Don't consider no rows an error because logs may not have been
			// uploaded yet.
			if sql.NoRowsInResultError(err) {
				return nil, nil
			}

			return nil, sql.Error(err)
		}

		return data, nil
	})
}
