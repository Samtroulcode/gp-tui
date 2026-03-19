package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.input = inputState{}
		m.setStatus("cancelled")
		return nil
	case "enter":
		return m.submitInput()
	case "backspace":
		if len(m.input.value) == 0 {
			return nil
		}

		runes := []rune(m.input.value)
		m.input.value = string(runes[:len(runes)-1])
		return nil
	}

	if len(msg.Runes) > 0 {
		m.input.value += string(msg.Runes)
	}

	return nil
}

func (m *Model) submitInput() tea.Cmd {
	value := strings.Trim(strings.TrimSpace(m.input.value), "/")
	if value == "" {
		m.setStatus("folder path is required")
		return nil
	}

	switch m.input.mode {
	case inputModeCreateDirectory:
		expanded := m.expandedDirectories()
		parent := parentDirectory(value)
		if parent != "" {
			expanded[parent] = true
		}

		m.input = inputState{}
		m.setStatus("creating folder %s", value)
		return createDirectoryCmd(m.service, value, expanded)
	}

	m.input = inputState{}
	return nil
}

func (m *Model) beginCreateDirectory() {
	base := m.currentDirectory()
	value := ""
	if base != "" {
		value = base + "/"
	}

	m.input = inputState{
		mode:   inputModeCreateDirectory,
		prompt: "New folder",
		value:  value,
	}
}
