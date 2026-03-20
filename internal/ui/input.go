package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleInput(msg tea.KeyMsg) tea.Cmd {
	if m.input.mode == inputModeDeleteEntries {
		return m.handleDeleteConfirmInput(msg)
	}
	if m.input.mode == inputModeSearch {
		return m.handleSearchInput(msg)
	}

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

func (m *Model) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.finishSearch(true)
		m.setStatus("cancelled")
		return nil
	case "enter":
		m.searchQuery = strings.TrimSpace(m.input.value)
		if m.searchQuery == "" {
			m.finishSearch(true)
			return nil
		}
		if m.currentNode() == nil {
			m.finishSearch(true)
			m.setStatus("no matching entry selected")
			return nil
		}

		m.finishSearchWithSelection()
		return nil
	case "backspace":
		if len(m.input.value) == 0 {
			return nil
		}

		runes := []rune(m.input.value)
		m.input.value = string(runes[:len(runes)-1])
		m.searchQuery = m.input.value
		m.applySearchFilter()
		return nil
	}

	if len(msg.Runes) > 0 {
		m.input.value += string(msg.Runes)
		m.searchQuery = m.input.value
		m.applySearchFilter()
	}

	return nil
}

func (m *Model) handleDeleteConfirmInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n":
		m.input = inputState{}
		m.setStatus("cancelled")
		return nil
	case "y", "enter":
		return m.submitInput()
	}

	return nil
}

func (m *Model) submitInput() tea.Cmd {
	if m.input.mode == inputModeDeleteEntries {
		paths := append([]string(nil), m.input.paths...)
		m.input = inputState{}
		m.setStatus("deleting %s", entryCountLabel(len(paths)))

		focusPath := m.currentDirectory()
		expanded := m.expandedStateForReload()
		if focusPath != "" {
			expanded[focusPath] = true
		}

		return deleteEntriesCmd(m.service, paths, focusPath, expanded)
	}

	value := strings.Trim(strings.TrimSpace(m.input.value), "/")
	if value == "" {
		m.setStatus("entry path is required")
		return nil
	}

	switch m.input.mode {
	case inputModeCreateEntry:
		m.input = inputState{}
		m.setStatus("creating entry %s", value)
		return createEntryCmd(m.service, value)
	}

	m.input = inputState{}
	return nil
}

func (m *Model) beginCreateEntry() {
	base := m.currentDirectory()
	value := ""
	if base != "" {
		value = base + "/"
	}

	m.input = inputState{
		mode:   inputModeCreateEntry,
		prompt: "New entry",
		value:  value,
	}
}
