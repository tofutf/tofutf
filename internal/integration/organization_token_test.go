package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/api"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/workspace"
)

// TestIntegration_OrganizationTokens demonstrates the use of an organization
// token to authenticate via the API.
func TestIntegration_OrganizationTokens(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	ot, token, err := daemon.Organizations.CreateToken(ctx, organization.CreateOrganizationTokenOptions{
		Organization: org.Name,
	})
	require.NoError(t, err)
	assert.Equal(t, org.Name, ot.Organization)

	apiClient, err := api.NewClient(api.Config{
		Address: daemon.System.Hostname(),
		Token:   string(token),
	})
	require.NoError(t, err)

	// create some workspaces and attempt to list them using client
	// authenticating with an organization token
	daemon.createWorkspace(t, ctx, org)
	daemon.createWorkspace(t, ctx, org)
	daemon.createWorkspace(t, ctx, org)

	wsClient := &workspace.Client{Client: apiClient}
	got, err := wsClient.List(ctx, workspace.ListOptions{
		Organization: internal.String(org.Name),
	})
	require.NoError(t, err)
	assert.Equal(t, 3, len(got.Items))

	// re-generate token
	_, _, err = daemon.Organizations.CreateToken(ctx, organization.CreateOrganizationTokenOptions{
		Organization: org.Name,
	})
	require.NoError(t, err)

	// access with previous token should now be refused
	_, err = wsClient.List(ctx, workspace.ListOptions{
		Organization: internal.String(org.Name),
	})
	require.Equal(t, internal.ErrUnauthorized, err)
}
