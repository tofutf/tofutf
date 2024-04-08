package organization

import (
	"context"
	"log/slog"

	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/rbac"
)

// Authorizer authorizes access to an organization
type Authorizer struct {
	Logger *slog.Logger
}

func (a *Authorizer) CanAccess(ctx context.Context, action rbac.Action, name string) (internal.Subject, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if internal.SkipAuthz(ctx) {
		return subj, nil
	}
	if subj.CanAccessOrganization(action, name) {
		return subj, nil
	}
	a.Logger.Error("unauthorized action", "organization", name, "action", action.String(), "subject", subj)
	return nil, internal.ErrAccessNotPermitted
}
