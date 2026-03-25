package gopass

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

type startupService interface {
	List(ctx context.Context) ([]string, error)
	ShowCommand(ctx context.Context, path string) *exec.Cmd
	SyncCommand(ctx context.Context) *exec.Cmd
}

// StartupIO describes the streams used during the interactive gopass startup flow.
type StartupIO struct {
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	UnlockOut io.Writer
}

// BootstrapStore unlocks the store when needed and runs an initial sync.
func BootstrapStore(ctx context.Context, service startupService, streams StartupIO) error {
	paths, err := service.List(ctx)
	if err != nil {
		return fmt.Errorf("list store: %w", err)
	}

	if len(paths) > 0 {
		if err := runInteractiveCommand(service.ShowCommand(ctx, paths[0]), "unlock store", streams.Stdin, streams.UnlockOut, streams.Stderr); err != nil {
			return fmt.Errorf("unlock store with %q: %w", paths[0], err)
		}
	}

	if err := runInteractiveCommand(service.SyncCommand(ctx), "sync store", streams.Stdin, streams.Stdout, streams.Stderr); err != nil {
		return fmt.Errorf("initial sync: %w", err)
	}

	return nil
}

func runInteractiveCommand(command *exec.Cmd, action string, stdin io.Reader, stdout, stderr io.Writer) error {
	if command == nil {
		return fmt.Errorf("%s: missing command", action)
	}

	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}

	return nil
}
