package repos

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func configureCancellation(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		kill := exec.CommandContext(ctx, "taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
		out, err := kill.CombinedOutput()
		if err != nil {
			return fmt.Errorf("canceling git process tree: %s: %w", strings.TrimSpace(string(out)), err)
		}

		return nil
	}
}
