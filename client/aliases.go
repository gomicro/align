package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

// AliasNames returns sorted global/system alias names, respecting Git's includes and environment.
func AliasNames(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "config", "--includes", "--null", "--show-scope", "--name-only", "--get-regexp", `^alias\.`)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 && stderr.Len() == 0 && ctx.Err() == nil {
			return nil, nil
		}

		detail := "reading git aliases"
		diagnostic := strings.TrimSpace(stderr.String())
		if diagnostic != "" {
			detail += ": " + diagnostic
		}

		return nil, fmt.Errorf("%s: %w", detail, err)
	}

	fields := strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")
	if len(fields)%2 != 0 {
		return nil, fmt.Errorf("reading git aliases: invalid scoped key output")
	}

	var names []string
	for i := 0; i < len(fields); i += 2 {
		scope := fields[i]
		if scope != "global" && scope != "system" {
			continue
		}

		name := strings.TrimPrefix(fields[i+1], "alias.")
		if name != "" {
			names = append(names, name)
		}
	}

	slices.Sort(names)
	return slices.Compact(names), nil
}
