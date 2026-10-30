// Package skills exports the agent guidance bundled with align.
package skills

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed align/SKILL.md
var files embed.FS

const skillPath = "align/SKILL.md"

// Dump exports the bundled skill, leaving differing content alone unless forced.
func Dump(ctx context.Context, outputDir string, force bool) (string, error) {
	if strings.TrimSpace(outputDir) == "" {
		return "", fmt.Errorf("--output-dir is required and must not be empty")
	}

	err := ctx.Err()
	if err != nil {
		return "", fmt.Errorf("exporting skill: %w", err)
	}

	content, err := files.ReadFile(skillPath)
	if err != nil {
		return "", fmt.Errorf("reading bundled skill: %w", err)
	}

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", fmt.Errorf("creating output directory: %w", err)
	}

	root, err := os.OpenRoot(outputDir)
	if err != nil {
		return "", fmt.Errorf("opening output directory: %w", err)
	}

	dumpErr := dump(ctx, root, content, force)
	closeErr := root.Close()
	err = errors.Join(dumpErr, closeErr)
	if err != nil {
		return "", fmt.Errorf("exporting skill: %w", err)
	}

	return filepath.Join(outputDir, filepath.FromSlash(skillPath)), nil
}

func dump(ctx context.Context, root *os.Root, content []byte, force bool) error {
	err := root.Mkdir("align", 0755)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("creating skill directory: %w", err)
	}

	info, err := root.Lstat("align")
	if err != nil {
		return fmt.Errorf("checking skill directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("skill directory must be a directory, not a symlink or file")
	}

	info, err = root.Lstat(skillPath)
	if errors.Is(err, fs.ErrNotExist) {
		return write(ctx, root, content, force)
	}

	if err != nil {
		return fmt.Errorf("checking existing skill: %w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("existing skill must be a regular file, not a symlink or directory")
	}

	existing, err := root.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("reading existing skill: %w", err)
	}

	if bytes.Equal(existing, content) {
		return nil
	}

	if !force {
		return fmt.Errorf("existing skill differs; use --force to replace it")
	}

	return write(ctx, root, content, force)
}

func write(ctx context.Context, root *os.Root, content []byte, force bool) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("writing skill: %w", err)
	}

	tempPath := "align/.SKILL-" + rand.Text() + ".tmp"
	file, err := root.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("creating temporary skill: %w", err)
	}

	n, writeErr := file.Write(content)
	closeErr := file.Close()
	err = errors.Join(writeErr, closeErr)
	if err == nil && n != len(content) {
		err = io.ErrShortWrite
	}

	if err == nil {
		err = install(ctx, root, tempPath, force)
	}

	removeErr := root.Remove(tempPath)
	if errors.Is(removeErr, fs.ErrNotExist) {
		removeErr = nil
	}

	if removeErr != nil {
		removeErr = fmt.Errorf("removing temporary skill: %w", removeErr)
	}

	err = errors.Join(err, removeErr)
	if err != nil {
		return fmt.Errorf("writing skill: %w", err)
	}

	return nil
}

func install(ctx context.Context, root *os.Root, tempPath string, force bool) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("installing skill: %w", err)
	}

	if force {
		err = root.Rename(tempPath, skillPath)
	} else {
		// Linking publishes the complete file without replacing a concurrent export.
		err = root.Link(tempPath, skillPath)
	}

	if err != nil {
		return fmt.Errorf("installing skill: %w", err)
	}

	return nil
}
