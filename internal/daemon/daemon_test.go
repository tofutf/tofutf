package daemon

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/xslog"
)

func TestDaemon_MissingSecretError(t *testing.T) {
	var missing *internal.MissingParameterError
	_, err := New(context.Background(), slog.New(&xslog.NoopHandler{}), Config{})
	require.True(t, errors.As(err, &missing))
}
