// Package xslog provides helpers related to logging.
package xslog

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*NoopHandler)(nil)

// NoopHandler is a slog handler that does not perform any logging.
type NoopHandler struct{}

// Enabled implements slog.Handler.
func (n *NoopHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

// Handle implements slog.Handler.
func (n *NoopHandler) Handle(context.Context, slog.Record) error {
	return nil
}

// WithAttrs implements slog.Handler.
func (n *NoopHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return n
}

// WithGroup implements slog.Handler.
func (n *NoopHandler) WithGroup(name string) slog.Handler {
	return n
}
