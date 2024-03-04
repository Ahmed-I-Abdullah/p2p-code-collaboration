package gitops

import (
	"fmt"
	"os/exec"
)

type Git struct {
	reposDir string
}

func New(reposDir string) *Git {
	return &Git{reposDir: reposDir}
}

func (g *Git) InitBare(repoName string) error {
	cmd := exec.Command("git", "init", "--bare", repoName)
	cmd.Dir = g.reposDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to init git: %w, output: %s", err, out)
	}
	return nil
}
