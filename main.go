package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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
	if len(paths) == 0 {
		return nil
	}

	command := service.ShowCommand(ctx, paths[0])
	var stderr bytes.Buffer
	command.Stdin = os.Stdin
	command.Stdout = io.Discard
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return fmt.Errorf("unlock store: %w", err)
		}

		return fmt.Errorf("unlock store: %s: %w", message, err)
	}

	return nil
}
