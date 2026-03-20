package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
	"gopass-tui/internal/ui"
)

func main() {
	service := gopass.NewService()
	if err := unlockStore(context.Background(), service); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	model, err := ui.NewModel(service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func unlockStore(ctx context.Context, service gopass.Service) error {
	paths, err := service.List(ctx)
	if err != nil {
		return err
	}
	if len(paths) > 0 {
		if err := runStartupCommand(service.ShowCommand(ctx, paths[0]), "unlock store", io.Discard); err != nil {
			return err
		}
	}

	if err := runStartupCommand(service.SyncCommand(ctx), "sync store", os.Stdout); err != nil {
		return err
	}

	return nil
}

func runStartupCommand(command *exec.Cmd, action string, stdout io.Writer) error {
	command.Stdin = os.Stdin
	command.Stdout = stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}

	return nil
}
