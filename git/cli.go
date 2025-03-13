package git

import (
	"net/url"
	"os/exec"
	"strings"

	"github.com/cashapp/hermit/errors"
	"github.com/cashapp/hermit/ui"
	"github.com/cashapp/hermit/util"
)

// CLIOperator implements Operator using git CLI commands
type CLIOperator struct {
	workDir string
	auth    Auth
}

// NewCLIOperator creates a new git operator that uses CLI commands
func NewCLIOperator(workDir string, auth Auth) *CLIOperator {
	return &CLIOperator{
		workDir: workDir,
		auth:    auth,
	}
}

// Clone implements Operator.Clone
func (c *CLIOperator) Clone(task *ui.Task, url string, opts CloneOpts) error {
	authenticatedURL, err := c.GetAuthenticatedURL(url)
	if err != nil {
		return errors.WithStack(err)
	}

	args := []string{"git", "clone"}
	if opts.Shallow {
		args = append(args, "--depth=1")
	}
	if opts.Reference != "" {
		args = append(args, "--branch", opts.Reference)
	}
	args = append(args, authenticatedURL, opts.TargetDir)

	return util.RunInDir(task, c.workDir, args...)
}

// Pull implements Operator.Pull
func (c *CLIOperator) Pull(task *ui.Task, repoDir string, opts PullOpts) error {
	if err := c.updateRemoteAuth(task, repoDir); err != nil {
		return err
	}

	args := []string{"git", "pull"}
	if opts.Force {
		args = append(args, "--force")
	}
	if opts.Reference != "" {
		args = append(args, "origin", opts.Reference)
	}

	return util.RunInDir(task, repoDir, args...)
}

// Fetch implements Operator.Fetch
func (c *CLIOperator) Fetch(task *ui.Task, repoDir string, opts FetchOpts) error {
	if err := c.updateRemoteAuth(task, repoDir); err != nil {
		return err
	}

	args := []string{"git", "fetch"}
	if opts.Force {
		args = append(args, "--force")
	}
	if opts.Reference != "" {
		args = append(args, "origin", opts.Reference)
	}

	return util.RunInDir(task, repoDir, args...)
}

// IsRepo implements Operator.IsRepo
func (c *CLIOperator) IsRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// GetAuthenticatedURL implements Operator.GetAuthenticatedURL
func (c *CLIOperator) GetAuthenticatedURL(url string) (string, error) {
	if c.auth.Token == "" || c.auth.ShouldAuth == nil {
		return url, nil
	}

	owner, repo, ok := parseGitHubURL(url)
	if !ok || !c.auth.ShouldAuth(owner, repo) {
		return url, nil
	}

	return addAuthToURL(url, c.auth.Token)
}

// updateRemoteAuth updates the remote URL with authentication if needed
func (c *CLIOperator) updateRemoteAuth(task *ui.Task, repoDir string) error {
	out, err := util.CaptureInDir(task, repoDir, "git", "remote", "get-url", "origin")
	if err != nil {
		return errors.WithStack(err)
	}

	currentURL := strings.TrimSpace(string(out))
	authenticatedURL, err := c.GetAuthenticatedURL(currentURL)
	if err != nil {
		return err
	}

	if authenticatedURL != currentURL {
		return util.RunInDir(task, repoDir, "git", "remote", "set-url", "origin", authenticatedURL)
	}

	return nil
}

// parseGitHubURL extracts owner and repo from a GitHub URL
func parseGitHubURL(urlStr string) (owner, repo string, ok bool) {
	// Handle HTTPS URLs
	if strings.HasPrefix(urlStr, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(urlStr, "https://github.com/"), "/")
		if len(parts) >= 2 {
			return parts[0], strings.TrimSuffix(parts[1], ".git"), true
		}
	}

	// Handle SSH URLs
	if strings.HasPrefix(urlStr, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(urlStr, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return parts[0], strings.TrimSuffix(parts[1], ".git"), true
		}
	}

	return "", "", false
}

// addAuthToURL adds authentication token to GitHub HTTPS URLs
func addAuthToURL(urlStr, token string) (string, error) {
	if !strings.HasPrefix(urlStr, "https://github.com/") {
		return urlStr, nil
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.WithStack(err)
	}

	u.User = url.UserPassword("x-access-token", token)
	return u.String(), nil
}
