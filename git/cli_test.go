package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/cashapp/hermit/ui"
)

func TestCLIOperator(t *testing.T) {
	l, _ := ui.NewForTesting()
	task := l.Task("test")

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test auth config
	auth := Auth{
		Token: "test-token",
		ShouldAuth: func(owner, repo string) bool {
			return owner == "cashapp"
		},
	}

	// Create operator
	op := NewCLIOperator(tmpDir, auth)

	t.Run("GetAuthenticatedURL", func(t *testing.T) {
		tests := []struct {
			name    string
			url     string
			want    string
			wantErr bool
		}{
			{
				name: "github https url with matching owner",
				url:  "https://github.com/cashapp/hermit.git",
				want: "https://x-access-token:test-token@github.com/cashapp/hermit.git",
			},
			{
				name: "github https url with non-matching owner",
				url:  "https://github.com/other/repo.git",
				want: "https://github.com/other/repo.git",
			},
			{
				name: "non-github url",
				url:  "https://gitlab.com/owner/repo.git",
				want: "https://gitlab.com/owner/repo.git",
			},
			{
				name: "ssh url",
				url:  "git@github.com:cashapp/hermit.git",
				want: "git@github.com:cashapp/hermit.git",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := op.GetAuthenticatedURL(tt.url)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("IsRepo", func(t *testing.T) {
		// Create a git repo
		repoDir := filepath.Join(tmpDir, "repo")
		err := os.MkdirAll(repoDir, 0750)
		assert.NoError(t, err)

		// Initialize git repo
		err = op.Clone(task, "https://github.com/cashapp/hermit.git", CloneOpts{
			TargetDir: repoDir,
			Shallow:   true,
		})
		if err != nil {
			t.Skip("Skipping test as unable to clone repository:", err)
		}

		// Test IsRepo
		assert.True(t, op.IsRepo(repoDir))
		assert.False(t, op.IsRepo(tmpDir))
	})
}
