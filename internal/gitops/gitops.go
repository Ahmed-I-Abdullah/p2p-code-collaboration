package gitops

import (
	"fmt"
	"os"
	"os/exec"
)

type Git struct {
	ReposDir string
}

func New(reposDir string) *Git {
	return &Git{ReposDir: reposDir}
}

func (g *Git) InitBare(repoName string) error {
	cmd := exec.Command("git", "init", "--bare", repoName)
	cmd.Dir = g.ReposDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to init git: %w, output: %s", err, out)
	}
	return nil
}

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
