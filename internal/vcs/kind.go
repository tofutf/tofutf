package vcs

const (
	GithubKind          Kind = "github"
	GitlabKind          Kind = "gitlab"
	BitbucketServerKind Kind = "bitbucketserver"
	GiteaKind           Kind = "gitea"
)

// Kind of vcs hosting provider
type Kind string

func KindPtr(k Kind) *Kind { return &k }
