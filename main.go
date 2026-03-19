package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
	"gopass-tui/internal/ui"
)

func main() {
	service := gopass.NewService()

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
