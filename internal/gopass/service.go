package gopass

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Service exposes the gopass operations needed by the TUI.
// Keeping this contract small makes the UI easier to test later.
type Service interface {
	List(ctx context.Context) ([]string, error)
	Show(ctx context.Context, path string) (string, error)
	ShowMasked(ctx context.Context, path string) (string, error)
	Copy(ctx context.Context, path string) error
	Move(ctx context.Context, sourcePath, destinationPath string) error
	Mkdir(ctx context.Context, path string) error
}

// CLIService implements Service by shelling out to the gopass binary.
type CLIService struct{}

// NewService returns the default gopass-backed service implementation.
func NewService() Service {
	return CLIService{}
}

// List returns all gopass entries as flat paths.
func (CLIService) List(ctx context.Context) ([]string, error) {
	output, err := runGopass(ctx, "ls", "--flat")
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}

	return paths, nil
}

// Show returns the full content of a gopass entry.
func (CLIService) Show(ctx context.Context, path string) (string, error) {
	output, err := runGopass(ctx, "show", path)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// ShowMasked returns the entry content with the first line hidden.
// This keeps the password secret while still showing additional metadata.
func (service CLIService) ShowMasked(ctx context.Context, path string) (string, error) {
	content, err := service.Show(ctx, path)
	if err != nil {
		return "", err
	}

	lines := strings.SplitN(content, "\n", 2)
	if len(lines) > 1 {
		return "********\n" + lines[1], nil
	}

	return "********", nil
}

// Copy delegates clipboard handling to gopass.
func (CLIService) Copy(ctx context.Context, path string) error {
	if _, err := runGopass(ctx, "show", "-c", path); err != nil {
		return err
	}

	return nil
}

// Move renames or relocates an entry through gopass.
func (CLIService) Move(ctx context.Context, sourcePath, destinationPath string) error {
	if _, err := runGopass(ctx, "mv", sourcePath, destinationPath); err != nil {
		return err
	}

	return nil
}

// Mkdir creates a directory in the gopass store.
func (CLIService) Mkdir(ctx context.Context, path string) error {
	if _, err := runGopass(ctx, "mkdir", path); err != nil {
		return err
	}

	return nil
}

func runGopass(ctx context.Context, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "gopass", args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message == "" {
			message = err.Error()
		}

		return nil, fmt.Errorf("gopass %s failed: %s: %w", strings.Join(args, " "), message, err)
	}

	return stdout.Bytes(), nil
}
