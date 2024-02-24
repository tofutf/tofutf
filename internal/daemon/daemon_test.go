package daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
)

func TestDaemon_MissingSecretError(t *testing.T) {
	var missing *internal.MissingParameterError
	_, err := New(context.Background(), logr.Discard(), Config{})
	require.True(t, errors.As(err, &missing))
}
