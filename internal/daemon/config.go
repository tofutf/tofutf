package daemon

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/agent"
	"github.com/tofutf/tofutf/internal/authenticator"
	"github.com/tofutf/tofutf/internal/configversion"
	"github.com/tofutf/tofutf/internal/inmem"
	"github.com/tofutf/tofutf/internal/tokens"
)

var ErrInvalidSecretLength = errors.New("secret must be 16 bytes in size")

// Config configures the otfd daemon. Descriptions of each field can be found in
// the flag definitions in ./cmd/otfd
type Config struct {
	AgentConfig *agent.Config
	CacheConfig *inmem.CacheConfig

	GithubHostname     string
	GithubURL          string
	GithubClientID     string
	GithubClientSecret string

	GitlabHostname     string
	GitlabURL          string
	GitlabClientID     string
	GitlabClientSecret string

	BitbucketServerHostname string
	BitbucketServerURL      string

	OIDC                         authenticator.OIDCConfig
	Secret                       []byte // 16-byte secret for signing URLs and encrypting payloads
	SiteToken                    string
	Host                         string
	WebhookHost                  string
	Address                      string
	Database                     string
	MaxConfigSize                int64
	SSL                          bool
	CertFile, KeyFile            string
	EnableRequestLogging         bool
	DevMode                      bool
	DisableScheduler             bool
	RestrictOrganizationCreation bool
	SiteAdmins                   []string
	SkipTLSVerification          bool
	// skip checks for latest terraform version
	DisableLatestChecker *bool

	tokens.GoogleIAPConfig
}

func ApplyDefaults(cfg *Config) {
	if cfg.AgentConfig == nil {
		cfg.AgentConfig = &agent.Config{
			Concurrency: agent.DefaultConcurrency,
		}
	}
	if cfg.CacheConfig == nil {
		cfg.CacheConfig = &inmem.CacheConfig{}
	}
	if cfg.MaxConfigSize == 0 {
		cfg.MaxConfigSize = configversion.DefaultConfigMaxSize
	}
}

func (cfg *Config) Valid() error {
	if cfg.Secret == nil {
		return &internal.MissingParameterError{Parameter: "secret"}
	}

	vcsURLs := map[string]string{
		"bitbucketserver": cfg.BitbucketServerURL,
		"github":          cfg.GithubURL,
		"gitlab":          cfg.GitlabURL,
	}

	for vcs, vcsURL := range vcsURLs {
		if vcsURL != "" {
			_, err := url.Parse(vcsURL)
			if err != nil {
				return internal.InvalidParameterError(fmt.Sprintf("invalid url for %s: %s", vcs, err.Error()))
			}
		}
	}

	if len(cfg.Secret) != 16 {
		return ErrInvalidSecretLength
	}

	return nil
}

// getURLFromHostnameAndURL calculates the base url for a vcs provider given a
// hostname, and url string. It is used to ensure compatability between the
// deprecated hostname configuration fields, and the current recommended url
// configuration fields.
func getURLFromHostnameAndURL(hostname, strURL string) *url.URL {
	if strURL == "" {
		if hostname == "" {
			return nil
		}

		return &url.URL{Scheme: "https", Host: hostname}
	}

	// this should never happen since url validation happens on the
	// configuration struct.
	parsed, err := url.Parse(strURL)
	if err != nil {
		panic(err)
	}

	return parsed
}
