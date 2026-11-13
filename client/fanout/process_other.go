//go:build !unix && !windows

package fanout

import "os/exec"

func configureCancellation(cmd *exec.Cmd) {
	// CommandContext kills the direct process on platforms without tree cancellation.
}
