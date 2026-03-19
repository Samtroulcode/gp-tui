package gopass

import (
	"fmt"
	"os/exec"
	"strings"
)

// Service exposes the gopass operations needed by the TUI.
// Keeping this contract small makes the UI easier to test later.
type Service interface {
	List() ([]string, error)
	Show(path string) (string, error)
	ShowMasked(path string) (string, error)
	Copy(path string) error
}

// CLIService implements Service by shelling out to the gopass binary.
type CLIService struct{}

// NewService returns the default gopass-backed service implementation.
func NewService() Service {
	return CLIService{}
}

// List returns all gopass entries as flat paths.
func (CLIService) List() ([]string, error) {
	command := exec.Command("gopass", "ls", "--flat")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("gopass ls --flat failed: %w", err)
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
func (CLIService) Show(path string) (string, error) {
	command := exec.Command("gopass", "show", path)
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("gopass show %q failed: %w", path, err)
	}

	return string(output), nil
}

// ShowMasked returns the entry content with the first line hidden.
// This keeps the password secret while still showing additional metadata.
func (service CLIService) ShowMasked(path string) (string, error) {
	content, err := service.Show(path)
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
func (CLIService) Copy(path string) error {
	command := exec.Command("gopass", "show", "-c", path)
	if err := command.Run(); err != nil {
		return fmt.Errorf("gopass show -c %q failed: %w", path, err)
	}

	return nil
}
