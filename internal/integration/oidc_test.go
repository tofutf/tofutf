package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/authenticator"
	"github.com/tofutf/tofutf/internal/daemon"
)

// NewOIDCIssuer creates an oidc issuer server and returns its url. For testing
// purposes.
func NewOIDCIssuer(t *testing.T, username, aud, name string) string {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	var u *url.URL
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		type discovery struct {
			Issuer            string   `json:"issuer"`
			Auth              string   `json:"authorization_endpoint"`
			Token             string   `json:"token_endpoint"`
			Keys              string   `json:"jwks_uri"`
			UserInfo          string   `json:"userinfo_endpoint"`
			DeviceEndpoint    string   `json:"device_authorization_endpoint"`
			GrantTypes        []string `json:"grant_types_supported"`
			ResponseTypes     []string `json:"response_types_supported"`
			Subjects          []string `json:"subject_types_supported"`
			IDTokenAlgs       []string `json:"id_token_signing_alg_values_supported"`
			CodeChallengeAlgs []string `json:"code_challenge_methods_supported"`
			Scopes            []string `json:"scopes_supported"`
			AuthMethods       []string `json:"token_endpoint_auth_methods_supported"`
			Claims            []string `json:"claims_supported"`
		}

		out, err := json.Marshal(&discovery{
			Issuer: u.String(),
			Auth:   absURL(u, "/login/oauth/authorize"),
			Token:  absURL(u, "/login/oauth/access_token"),
			Keys:   absURL(u, "/keys"),
		})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out) //nolint:errcheck
	})
	// auth endpoint
	mux.HandleFunc("/login/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := url.Values{}
		// send same state back to client
		q.Add("state", r.URL.Query().Get("state"))
		// generate any old code; the token endpoint will receive it later and
		// disregard it.
		q.Add("code", internal.GenerateRandomString(10))

		referrer, err := url.Parse(r.Referer())
		require.NoError(t, err)

		// construct redirect url
		callback := url.URL{
			Scheme:   referrer.Scheme,
			Host:     referrer.Host,
			Path:     fmt.Sprintf("/oauth/%s/callback", name),
			RawQuery: q.Encode(),
		}

		http.Redirect(w, r, callback.String(), http.StatusFound)
	})
	// token endpoint
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		token := fakeIDToken(t, username, aud, u.String(), priv)

		out, err := json.Marshal(struct {
			AccessToken string `json:"access_token"`
			IDToken     string `json:"id_token"`
		}{
			AccessToken: "stub_token",
			IDToken:     token,
		})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out) //nolint:errcheck
	})
	// keyset endpoint
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		key, err := jwk.FromRaw(priv.Public())
		set := jwk.NewSet()
		set.AddKey(key) //nolint:errcheck
		require.NoError(t, err)
		out, err := json.Marshal(set)
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out) //nolint:errcheck
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("oidc issuer received request for non-existent path: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)

	u, err = url.Parse(srv.URL)
	require.NoError(t, err)

	return u.String()
}

func absURL(u *url.URL, path string) string {
	u2 := *u
	u2.Path = path
	return u2.String()
}

func fakeIDToken(t *testing.T, name, aud, issuer string, key *rsa.PrivateKey) string {
	token, err := jwt.NewBuilder().
		Claim("name", name).
		Audience([]string{aud}).
		Issuer(issuer).
		Expiration(time.Now().Add(time.Minute)).
		Build()
	require.NoError(t, err)

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, key))
	require.NoError(t, err)
	return string(signed)
}

// TestIntegration_OIDC demonstrates logging in using OIDC
func TestIntegration_OIDC(t *testing.T) {
	integrationTest(t)

	// Start daemon with a stub github server populated with a user.
	cfg := config{
		Config: daemon.Config{
			OIDC: authenticator.OIDCConfig{
				Name:                "google",
				IssuerURL:           NewOIDCIssuer(t, "bobby", "stub-client-id", "google"),
				ClientID:            "stub-client-id",
				ClientSecret:        "stub-client-secret",
				SkipTLSVerification: true,
				UsernameClaim:       string(authenticator.DefaultUsernameClaim),
			},
		},
	}

	svc, _, _ := setup(t, &cfg)

	browser.Run(t, nil, chromedp.Tasks{
		// go to login page
		chromedp.Navigate("https://" + svc.System.Hostname() + "/login"),
		screenshot(t, "oidc_login_button"),
		// login
		chromedp.Click("a#login-button-google"),
		screenshot(t),
		// check login confirmation message
		matchText(t, "#content > p", "You are logged in as bobby", chromedp.ByQuery),
	})
}
