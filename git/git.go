package git

import (
	"fmt"
	"os/exec"
)

type Git struct {
}

func NewClient() (*Git, error) {
	//TODO: validate git exists on system or return an error

	return &Git{}, nil
}

func (g *Git) Run(subcmd string, args ...string) error {
	cmdargs := make([]string, len(args)+1)
	cmdargs = append(cmdargs, subcmd)
	cmdargs = append(cmdargs, args...)

	cmd := exec.Command("git", cmdargs...)

	// TODO: Allow for optional stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}
