package vcs

const (
	GithubKind      Kind = "github"
	GitlabKind      Kind = "gitlab"
	BitbucketServer Kind = "bitbucketserver"
)

// Kind of vcs hosting provider
type Kind string

func KindPtr(k Kind) *Kind { return &k }
