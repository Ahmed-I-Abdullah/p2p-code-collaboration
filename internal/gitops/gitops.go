package gitops

import (
	"fmt"
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
