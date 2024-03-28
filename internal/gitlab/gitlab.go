// Package gitlab provides gitlab related code
package gitlab

import (
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

const (
	// DefaultHostname is the default host for the gitlab vcs provider.
	//
	// Deprecated: use DefaultURL instead.
	DefaultHostname string = "gitlab.com"

	DefaultURL string = "https://gitlab.com"
)

var (
	OAuthEndpoint = oauth2gitlab.Endpoint
	OAuthScopes   = []string{"read_user", "read_api"}
)
