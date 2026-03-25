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
	ShowCommand(ctx context.Context, path string) *exec.Cmd
	SyncCommand(ctx context.Context) *exec.Cmd
	Show(ctx context.Context, path string) (string, error)
	ShowMasked(ctx context.Context, path string) (string, error)
	EditCommand(ctx context.Context, path string) *exec.Cmd
	CreateCommand(ctx context.Context, path string) *exec.Cmd
	GenerateCommand(ctx context.Context, request GenerateRequest) (*exec.Cmd, error)
	Copy(ctx context.Context, path string) error
	Delete(ctx context.Context, path string) error
	Move(ctx context.Context, sourcePath, destinationPath string) error
}

// GenerateRequest describes a gopass password generation request.
type GenerateRequest struct {
	Path              string
	Key               string
	Length            int
	Clip              bool
	Print             bool
	Force             bool
	Edit              bool
	Symbols           bool
	Generator         string
	Strict            bool
	ForceRegen        bool
	Separator         string
	Language          string
	CommitMessage     string
	InteractiveCommit bool
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

// ShowCommand returns an interactive gopass show process for an entry.
func (CLIService) ShowCommand(ctx context.Context, path string) *exec.Cmd {
	return exec.CommandContext(ctx, "gopass", "show", "--", path)
}

// Show returns the full content of a gopass entry.
func (CLIService) Show(ctx context.Context, path string) (string, error) {
	output, err := runGopass(ctx, "show", "--", path)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// SyncCommand returns an interactive gopass sync process.
func (CLIService) SyncCommand(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "gopass", "sync")
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

// EditCommand returns an interactive gopass edit process for an entry.
func (CLIService) EditCommand(ctx context.Context, path string) *exec.Cmd {
	return exec.CommandContext(ctx, "gopass", "edit", "--", path)
}

// CreateCommand returns an interactive gopass edit process for a new entry.
func (CLIService) CreateCommand(ctx context.Context, path string) *exec.Cmd {
	return exec.CommandContext(ctx, "gopass", "edit", "--create", "--", path)
}

// GenerateCommand returns a gopass generate process for a password generation request.
func (service CLIService) GenerateCommand(ctx context.Context, request GenerateRequest) (*exec.Cmd, error) {
	args, err := service.generateArgs(request)
	if err != nil {
		return nil, err
	}

	return exec.CommandContext(ctx, "gopass", args...), nil
}

// Copy delegates clipboard handling to gopass.
func (CLIService) Copy(ctx context.Context, path string) error {
	if _, err := runGopass(ctx, "show", "-c", "--", path); err != nil {
		return err
	}

	return nil
}

// Delete removes an entry through gopass.
func (CLIService) Delete(ctx context.Context, path string) error {
	if _, err := runGopass(ctx, "rm", "-f", "--", path); err != nil {
		return err
	}

	return nil
}

// Move renames or relocates an entry through gopass.
func (CLIService) Move(ctx context.Context, sourcePath, destinationPath string) error {
	if _, err := runGopass(ctx, "mv", "--", sourcePath, destinationPath); err != nil {
		return err
	}

	return nil
}

func (CLIService) generateArgs(request GenerateRequest) ([]string, error) {
	path := strings.TrimSpace(request.Path)
	if path == "" {
		return nil, fmt.Errorf("entry path is required")
	}
	if request.Length <= 0 {
		return nil, fmt.Errorf("password length must be positive")
	}

	generator, err := normalizeGenerator(request.Generator)
	if err != nil {
		return nil, err
	}

	language, err := normalizeLanguage(request.Language)
	if err != nil {
		return nil, err
	}

	args := []string{"generate"}
	args = appendEnabledArgs(args,
		flagArgs{enabled: request.Clip, args: []string{"--clip"}},
		flagArgs{enabled: request.Print, args: []string{"--print"}},
		flagArgs{enabled: request.Force, args: []string{"--force"}},
		flagArgs{enabled: request.Edit, args: []string{"--edit"}},
		flagArgs{enabled: request.Symbols, args: []string{"--symbols"}},
	)
	args = append(args, "--generator", generator)
	args = appendEnabledArgs(args,
		flagArgs{enabled: request.Strict, args: []string{"--strict"}},
		flagArgs{enabled: request.ForceRegen, args: []string{"--force-regen"}},
	)
	if request.Separator != "" {
		args = append(args, "--sep", request.Separator)
	}
	args = append(args, "--lang", language)
	if request.CommitMessage != "" {
		args = append(args, "--commit-message", request.CommitMessage)
	}
	args = appendEnabledArgs(args, flagArgs{enabled: request.InteractiveCommit, args: []string{"--interactive-commit"}})

	args = append(args, "--", path)
	if key := strings.TrimSpace(request.Key); key != "" {
		args = append(args, key)
	}
	args = append(args, fmt.Sprintf("%d", request.Length))

	return args, nil
}

type flagArgs struct {
	enabled bool
	args    []string
}

func appendEnabledArgs(base []string, flags ...flagArgs) []string {
	for _, flag := range flags {
		if flag.enabled {
			base = append(base, flag.args...)
		}
	}

	return base
}

func normalizeGenerator(generator string) (string, error) {
	value := strings.TrimSpace(generator)
	if value == "" {
		return "cryptic", nil
	}

	switch value {
	case "cryptic", "memorable", "xkcd", "external":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported generator %q", value)
	}
}

func normalizeLanguage(language string) (string, error) {
	value := strings.TrimSpace(language)
	if value == "" {
		return "en", nil
	}

	switch value {
	case "en", "de":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported language %q", value)
	}
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
			message = err.Error()
		}

		return nil, fmt.Errorf("gopass %s failed: %s: %w", strings.Join(args, " "), message, err)
	}

	return stdout.Bytes(), nil
}
