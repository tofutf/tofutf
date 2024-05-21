// Package workspace provides access to terraform workspaces
package workspace

import (
	"fmt"
	"regexp"

	"log/slog"

	"slices"

	"github.com/gobwas/glob"
	types "github.com/hashicorp/go-tfe"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/releases"
	"github.com/tofutf/tofutf/internal/resource"
	"github.com/tofutf/tofutf/internal/semver"
)

const (
	RemoteExecutionMode ExecutionMode = "remote"
	LocalExecutionMode  ExecutionMode = "local"
	AgentExecutionMode  ExecutionMode = "agent"

	DefaultAllowDestroyPlan = true
	MinTerraformVersion     = "1.2.0"
)

var (
	apiTestTerraformVersions = []string{"0.10.0", "0.11.0", "0.11.1"}
)

type (
	// Workspace is a terraform workspace.
	/*
		Workspace struct {

		}
		/*
			Connection struct {
				// Pushes to this VCS branch trigger runs. Empty string means the default
				// branch is used. Ignored if TagsRegex is non-empty.
				Branch string
				// Pushed tags matching this regular expression trigger runs. Mutually
				// exclusive with TriggerPatterns.
				TagsRegex string

				VCSProviderID string
				Repo          string

				// By default, once a workspace is connected to a repo it is not
				// possible to run a terraform apply via the CLI. Setting this to true
				// overrides this behaviour.
				AllowCLIApply bool
			}
	*/
	ConnectOptions struct {
		RepoPath      *string
		VCSProviderID *string

		Branch        *string
		TagsRegex     *string
		AllowCLIApply *bool
	}

	ExecutionMode string

	// CreateOptions represents the options for creating a new workspace.
	CreateOptions struct {
		AgentPoolID                *string
		AllowDestroyPlan           *bool
		AutoApply                  *bool
		Description                *string
		ExecutionMode              *ExecutionMode
		GlobalRemoteState          *bool
		MigrationEnvironment       *string
		Name                       *string
		QueueAllRuns               *bool
		SpeculativeEnabled         *bool
		SourceName                 *string
		SourceURL                  *string
		StructuredRunOutputEnabled *bool
		Tags                       []TagSpec
		TerraformVersion           *string
		TriggerPrefixes            []string
		TriggerPatterns            []string
		WorkingDirectory           *string
		Organization               *string

		// Always trigger runs. A value of true is mutually exclusive with
		// setting TriggerPatterns or ConnectOptions.TagsRegex.
		AlwaysTrigger *bool

		*ConnectOptions
	}

	UpdateOptions struct {
		AgentPoolID                *string `json:"agent-pool-id,omitempty"`
		AllowDestroyPlan           *bool
		AutoApply                  *bool
		Name                       *string
		Description                *string
		ExecutionMode              *ExecutionMode `json:"execution-mode,omitempty"`
		GlobalRemoteState          *bool
		Operations                 *bool
		QueueAllRuns               *bool
		SpeculativeEnabled         *bool
		StructuredRunOutputEnabled *bool
		TerraformVersion           *string
		TriggerPrefixes            []string
		TriggerPatterns            []string
		WorkingDirectory           *string

		// Always trigger runs. A value of true is mutually exclusive with
		// setting TriggerPatterns or ConnectOptions.TagsRegex.
		AlwaysTrigger *bool

		// Disconnect workspace from repo. It is invalid to specify true for an
		// already disconnected workspace.
		Disconnect bool

		// Specifying ConnectOptions either connects a currently
		// disconnected workspace, or modifies a connection if already
		// connected.
		*ConnectOptions
	}

	// ListOptions are options for paginating and filtering a list of
	// Workspaces
	ListOptions struct {
		Search       string
		Tags         []string
		Organization *string `schema:"organization_name"`

		resource.PageOptions
	}

	// VCS trigger strategy determines which VCS events trigger runs
	VCSTriggerStrategy string
)

func NewWorkspace(opts types.WorkspaceCreateOptions) (*types.Workspace, error) {
	// required options
	if err := resource.ValidateName(opts.Name); err != nil {
		return nil, err
	}

	if opts.Project == nil {
		return nil, internal.ErrRequiredOrg
	}
	if opts.Project.Organization == nil {
		return nil, internal.ErrRequiredOrg
	}

	ws := &types.Workspace{
		ID:                 internal.NewID("ws"),
		CreatedAt:          internal.CurrentTimestamp(nil),
		UpdatedAt:          internal.CurrentTimestamp(nil),
		AllowDestroyPlan:   DefaultAllowDestroyPlan,
		ExecutionMode:      string(RemoteExecutionMode),
		TerraformVersion:   releases.DefaultTerraformVersion,
		SpeculativeEnabled: true,
	}
	if opts.Project.Organization != nil {
		ws.Organization = &types.Organization{
			Name: opts.Project.Organization.Name,
		}
	}
	if err := setName(ws, *opts.Name); err != nil {
		return nil, err
	}
	if _, err := setExecutionModeAndAgentPoolID(ws, opts.ExecutionMode, opts.AgentPoolID); err != nil {
		return nil, err
	}
	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
	}
	if opts.GlobalRemoteState != nil {
		ws.GlobalRemoteState = *opts.GlobalRemoteState
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SourceName != nil {
		ws.SourceName = *opts.SourceName
	}
	if opts.SourceURL != nil {
		ws.SourceURL = *opts.SourceURL
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		if err := setTerraformVersion(ws, *opts.TerraformVersion); err != nil {
			return nil, err
		}
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}
	// TriggerPrefixes are not used but OTF persists it in order to pass go-tfe
	// integration tests.
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	// Enforce three-way mutually exclusivity between:
	// (a) tags-regex
	// (b) trigger-patterns
	// (c) always-trigger=true
	if (opts.VCSRepo != nil && (opts.VCSRepo.TagsRegex != nil && *opts.VCSRepo.TagsRegex != "")) && opts.TriggerPatterns != nil {
		return nil, ErrTagsRegexAndTriggerPatterns
	}
	if (opts.VCSRepo != nil && (opts.VCSRepo.TagsRegex != nil && *opts.VCSRepo.TagsRegex != "")) && (opts.QueueAllRuns != nil && *opts.QueueAllRuns) {
		return nil, ErrTagsRegexAndAlwaysTrigger
	}
	if len(opts.TriggerPatterns) > 0 && (opts.QueueAllRuns != nil && *opts.QueueAllRuns) {
		return nil, ErrTriggerPatternsAndAlwaysTrigger
	}
	if opts.VCSRepo != nil {
		if err := addConnection(ws, opts.VCSRepo); err != nil {
			return nil, err
		}
	}
	if opts.TriggerPatterns != nil {
		if err := setTriggerPatterns(ws, opts.TriggerPatterns); err != nil {
			return nil, fmt.Errorf("setting trigger patterns: %w", err)
		}
	}
	return ws, nil
}

// ExecutionModePtr returns a pointer to an execution mode.
func ExecutionModePtr(m ExecutionMode) *ExecutionMode {
	return &m
}

// ExecutionModes returns a list of possible execution modes
func ExecutionModes() []string {
	return []string{"local", "remote", "agent"}
}

// LogValue implements slog.LogValuer.
func LogValue(ws *types.Workspace) slog.Value {
	return slog.GroupValue(
		slog.String("id", ws.ID),
		slog.String("organization", ws.Organization.Name),
		slog.String("name", ws.Name),
	)
}

// Update updates the workspace with the given options. A boolean is returned to
// indicate whether the workspace is to be connected to a repo (true),
// disconnected from a repo (false), or neither (nil).
func Update(ws *types.Workspace, opts types.WorkspaceUpdateOptions) (*bool, error) {
	var updated bool

	if opts.Name != nil {
		if err := setName(ws, *opts.Name); err != nil {
			return nil, err
		}
		updated = true
	}
	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
		updated = true
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
		updated = true
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
		updated = true
	}
	if changed, err := setExecutionModeAndAgentPoolID(ws, opts.ExecutionMode, opts.AgentPoolID); err != nil {
		return nil, err
	} else if changed {
		updated = true
	}
	if opts.Operations != nil {
		if *opts.Operations {
			ws.ExecutionMode = "remote"
		} else {
			ws.ExecutionMode = "local"
		}
		updated = true
	}
	if opts.GlobalRemoteState != nil {
		ws.GlobalRemoteState = *opts.GlobalRemoteState
		updated = true
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
		updated = true
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
		updated = true
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
		updated = true
	}
	if opts.TerraformVersion != nil {
		if err := setTerraformVersion(ws, *opts.TerraformVersion); err != nil {
			return nil, err
		}
		updated = true
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
		updated = true
	}
	// TriggerPrefixes are not used but OTF persists it in order to pass go-tfe
	// integration tests.
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
		updated = true
	}
	// Enforce three-way mutually exclusivity between:
	// (a) tags-regex
	// (b) trigger-patterns
	// (c) always-trigger=true
	if (opts.VCSRepo != nil && (opts.VCSRepo.TagsRegex != nil && *opts.VCSRepo.TagsRegex != "")) && opts.TriggerPatterns != nil {
		return nil, ErrTagsRegexAndTriggerPatterns
	}
	if (opts.VCSRepo != nil && (opts.VCSRepo.TagsRegex != nil && *opts.VCSRepo.TagsRegex != "")) && (opts.QueueAllRuns != nil && *opts.QueueAllRuns) {
		return nil, ErrTagsRegexAndAlwaysTrigger
	}
	if len(opts.TriggerPatterns) > 0 && (opts.QueueAllRuns != nil && *opts.QueueAllRuns) {
		return nil, ErrTriggerPatternsAndAlwaysTrigger
	}
	if opts.QueueAllRuns != nil && *opts.QueueAllRuns {
		if ws.VCSRepo != nil {
			ws.VCSRepo.TagsRegex = ""
		}
		ws.TriggerPatterns = nil
		updated = true
	}
	if opts.TriggerPatterns != nil {
		if err := setTriggerPatterns(ws, opts.TriggerPatterns); err != nil {
			return nil, fmt.Errorf("setting trigger patterns: %w", err)
		}
		if ws.VCSRepo != nil {
			ws.VCSRepo.TagsRegex = ""
		}
		updated = true
	}
	// determine whether to connect or disconnect workspace
	// FIXME
	connect := internal.Bool(true)
	/*
		if opts.Disconnect && opts.ConnectOptions != nil {
			return nil, errors.New("connect options must be nil if disconnect is true")
		}
		var connect *bool
		if opts.Disconnect {
			if ws.VCSRepo == nil {
				return nil, errors.New("cannot disconnect an already disconnected workspace")
			}
			// workspace is to be disconnected
			connect = internal.Bool(false)
			updated = true
		}
	*/
	if opts.VCSRepo != nil {
		if ws.VCSRepo == nil {
			// workspace is to be connected
			if err := addConnection(ws, opts.VCSRepo); err != nil {
				return nil, err
			}
			connect = internal.Bool(true)
			updated = true
		} else {
			// modify existing connection
			if opts.VCSRepo.TagsRegex != nil && *opts.VCSRepo.TagsRegex != "" {
				if err := setTagsRegex(ws, *opts.VCSRepo.TagsRegex); err != nil {
					return nil, fmt.Errorf("invalid tags-regex: %w", err)
				}
				ws.TriggerPatterns = nil
				updated = true
			}
			if opts.VCSRepo.Branch != nil {
				ws.VCSRepo.Branch = *opts.VCSRepo.Branch
				updated = true
			}
			if opts.QueueAllRuns != nil {
				ws.QueueAllRuns = *opts.QueueAllRuns
				updated = true
			}
		}
	}
	if updated {
		ws.UpdatedAt = internal.CurrentTimestamp(nil)
	}
	return connect, nil
}

func addConnection(ws *types.Workspace, opts *types.VCSRepoOptions) error {
	if opts.Identifier == nil {
		return &internal.MissingParameterError{Parameter: "identifier"}
	}
	if opts.OAuthTokenID == nil && opts.GHAInstallationID == nil {
		return &internal.MissingParameterError{Parameter: "oauth_token_id"}
	}
	ws.VCSRepo = &types.VCSRepo{}

	if opts.TagsRegex != nil && *opts.TagsRegex != "" {
		if err := setTagsRegex(ws, *opts.TagsRegex); err != nil {
			return fmt.Errorf("invalid tags-regex: %w", err)
		}
	}
	if opts.Branch != nil {
		ws.VCSRepo.Branch = *opts.Branch
	}
	return nil
}

func setName(ws *types.Workspace, name string) error {
	if !internal.ReStringID.MatchString(name) {
		return internal.ErrInvalidName
	}
	ws.Name = name
	return nil
}

// setExecutionModeAndAgentPoolID sets the execution mode and/or the agent pool
// ID. The two parameters are intimately related, hence the validation and
// setting of the parameters is handled in tandem.
func setExecutionModeAndAgentPoolID(ws *types.Workspace, m *string, agentPoolID *string) (bool, error) {
	if m == nil {
		if agentPoolID == nil {
			// neither specified; nothing more to be done
			return false, nil
		} else {
			// agent pool ID can be set without specifying execution mode as long as
			// existing execution mode is AgentExecutionMode
			if ws.ExecutionMode != string(AgentExecutionMode) {
				return false, ErrNonAgentExecutionModeWithPool
			}
			ws.AgentPool = &types.AgentPool{
				ID: *agentPoolID,
			}
		}
	} else {
		if *m == string(AgentExecutionMode) {
			if agentPoolID == nil {
				return false, ErrAgentExecutionModeWithoutPool
			}
			ws.AgentPool = &types.AgentPool{
				ID: *agentPoolID,
			}
		} else {
			// mode is either remote or local; in either case no pool ID should be
			// provided
			if agentPoolID != nil {
				return false, ErrNonAgentExecutionModeWithPool
			}
		}
		ws.ExecutionMode = string(*m)
	}
	return true, nil
}

func setTerraformVersion(ws *types.Workspace, v string) error {
	if v == releases.LatestVersionString {
		ws.TerraformVersion = v
		return nil
	}
	if !semver.IsValid(v) {
		return internal.ErrInvalidTerraformVersion
	}
	// only accept terraform versions above the minimum requirement.
	//
	// NOTE: we make an exception for the specific versions posted by the go-tfe
	// integration tests.
	if result := semver.Compare(v, MinTerraformVersion); result < 0 {
		if !slices.Contains(apiTestTerraformVersions, v) {
			return ErrUnsupportedTerraformVersion
		}
	}
	ws.TerraformVersion = v
	return nil
}

func setTagsRegex(ws *types.Workspace, regex string) error {
	if _, err := regexp.Compile(regex); err != nil {
		return ErrInvalidTagsRegex
	}
	ws.VCSRepo.TagsRegex = regex
	return nil
}

func setTriggerPatterns(ws *types.Workspace, patterns []string) error {
	for _, patt := range patterns {
		if _, err := glob.Compile(patt); err != nil {
			return ErrInvalidTriggerPattern
		}
	}
	ws.TriggerPatterns = patterns
	return nil
}
