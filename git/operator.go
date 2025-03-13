package git

import (
	"github.com/cashapp/hermit/ui"
)

// Operator defines the interface for git operations
type Operator interface {
	// Clone a repository
	Clone(task *ui.Task, url string, opts CloneOpts) error

	// Pull updates from remote
	Pull(task *ui.Task, repoDir string, opts PullOpts) error

	// Fetch from remote without merging
	Fetch(task *ui.Task, repoDir string, opts FetchOpts) error

	// IsRepo checks if directory is a git repository
	IsRepo(dir string) bool

	// GetAuthenticatedURL returns URL with auth if needed
	GetAuthenticatedURL(url string) (string, error)
}

// CloneOpts configures clone operation
type CloneOpts struct {
	Shallow   bool
	Reference string
	TargetDir string
}

// PullOpts configures pull operation
type PullOpts struct {
	Force     bool
	Reference string
}

// FetchOpts configures fetch operation
type FetchOpts struct {
	Force     bool
	Reference string
}

// Auth represents authentication configuration for git operations
type Auth struct {
	// Token is the authentication token (e.g. GitHub token)
	Token string
	// ShouldAuth determines if authentication should be used for a given owner/repo
	ShouldAuth func(owner, repo string) bool
}
