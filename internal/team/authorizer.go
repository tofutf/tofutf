package team

import (
	"context"
	"log/slog"

	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/rbac"
)

// authorizer authorizes access to a team
type authorizer struct {
	Logger *slog.Logger
}

func (a *authorizer) CanAccess(ctx context.Context, action rbac.Action, teamID string) (internal.Subject, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if internal.SkipAuthz(ctx) {
		return subj, nil
	}
	if subj.CanAccessTeam(action, teamID) {
		return subj, nil
	}
	a.Logger.Error("unauthorized action", "team_id", teamID, "action", action.String(), "subject", subj)
	return nil, internal.ErrAccessNotPermitted
}
