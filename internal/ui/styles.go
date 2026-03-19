package ui

import "github.com/charmbracelet/lipgloss"

var (
	styleDir = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true)

	styleEntry = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))

	styleCursor = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f5c2e7")).
			Bold(true)

	styleSelected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1")).
			Bold(true)

	styleCut = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f9e2af")).
			Bold(true)

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70"))

	styleStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fab387"))

	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cba6f7")).
			Bold(true).
			Padding(0, 1)

	stylePreview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9399b2")).
			Padding(0, 2)
)
