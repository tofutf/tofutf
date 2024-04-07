package internal

import (
	"context"
	"log/slog"

	"github.com/tofutf/tofutf/internal/rbac"
)

// SiteAuthorizer authorizes access to site-wide actions
type SiteAuthorizer struct {
	Logger *slog.Logger
}

func (a *SiteAuthorizer) CanAccess(ctx context.Context, action rbac.Action, _ string) (Subject, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subj.CanAccessSite(action) {
		return subj, nil
	}
	a.Logger.Error("unauthorized action", "action", action, "subject", subj)
	return nil, ErrAccessNotPermitted
}
