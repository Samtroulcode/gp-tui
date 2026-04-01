package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorSurface      = lipgloss.Color("#1e1e2e")
	colorSurfaceAlt   = lipgloss.Color("#181825")
	colorBorder       = lipgloss.Color("#313244")
	colorBorderActive = lipgloss.Color("#89b4fa")
	colorText         = lipgloss.Color("#cdd6f4")
	colorMuted        = lipgloss.Color("#7f849c")
	colorAccent       = lipgloss.Color("#cba6f7")
	colorSuccess      = lipgloss.Color("#a6e3a1")
	colorWarn         = lipgloss.Color("#f9e2af")
	colorCursor       = lipgloss.Color("#f5c2e7")
	colorPreview      = lipgloss.Color("#bac2de")

	styleApp = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorSurface)

	stylePanel = lipgloss.NewStyle().
			Background(colorSurface).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	stylePanelTitle = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	stylePanelMeta = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleSearchLabel = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleSearchBox = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorSurfaceAlt).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	styleSearchBoxActive = styleSearchBox.Copy().
				BorderForeground(colorBorderActive)

	styleStatusPanel = stylePanel.Copy().
				BorderForeground(colorAccent)

	styleModal = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorSurfaceAlt).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	styleDir = lipgloss.NewStyle().
			Foreground(colorBorderActive).
			Bold(true)

	styleEntry = lipgloss.NewStyle().
			Foreground(colorText)

	styleCursor = lipgloss.NewStyle().
			Foreground(colorCursor).
			Bold(true)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	styleCut = lipgloss.NewStyle().
			Foreground(colorWarn).
			Bold(true)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleHelpPanel = lipgloss.NewStyle().
			Foreground(colorPreview).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	styleStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fab387"))

	styleTitle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 1)

	stylePreview = lipgloss.NewStyle().
			Foreground(colorPreview).
			Padding(0, 2)

	stylePlaceholder = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleStatusLabel = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleStatusValue = lipgloss.NewStyle().
				Foreground(colorText)
)
