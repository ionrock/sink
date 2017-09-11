package repo

import (
	"os"
	"os/exec"
	"path"
	"strings"
)

type Git struct {
	Remote string
}

func (git *Git) path() string {
	return strings.TrimSuffix(path.Base(git.Remote), ".git")
}

func (git *Git) Clone() error {
	_, err := os.Stat(git.path())

	if err == nil {
		return err
	}

	if !os.IsNotExist(err) {
		return err
	}

	cmd := exec.Command("git", "clone", git.Remote, git.path())

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
