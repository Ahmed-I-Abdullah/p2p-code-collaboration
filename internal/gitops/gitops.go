// Package gitops provides utilities for managing Git repositories and operations
package gitops

import (
	"fmt"
	"os"
	"os/exec"
)

// Git represents operations related to Git repositories
type Git struct {
	// ReposDir is the directory where Git repositories are stored (from which deamon serves repos)
	ReposDir string
}

// New creates a new Git instance with the provided repositories directory
func New(reposDir string) *Git {
	return &Git{ReposDir: reposDir}
}

// InitBare initializes a new bare Git repository with the specified repoName
func (g *Git) InitBare(repoName string) error {
	cmd := exec.Command("git", "init", "--bare", repoName)
	cmd.Dir = g.ReposDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize a bare Git repository: %w, output: %s", err, out)
	}
	return nil
}

// PushToPeer pushes changes from a local repository to a peer's repository
// repoName specifies the name of the local repository
// peerGitAddress specifies the Git address of the peer's repository
func (g *Git) PushToPeer(repoName, peerGitAddress string) error {
	repoPath := fmt.Sprintf("%s/%s", g.ReposDir, repoName)

	cmd := exec.Command("git", "push", peerGitAddress, "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = repoPath

	if err := cmd.Run(); err != nil {
		fmt.Errorf("Failed to execute a git push with error: %v", err)
		return fmt.Errorf("failed to push changes: %v", err)
	}

	return nil
}
