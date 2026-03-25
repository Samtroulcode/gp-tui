package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleInput(msg tea.KeyMsg) tea.Cmd {
	switch m.input.mode {
	case inputModeDeleteEntries:
		return m.handleDeleteConfirmInput(msg)
	case inputModeSearch:
		return m.handleSearchInput(msg)
	case inputModeGenerateWizard:
		return m.handleGenerateWizardInput(msg)
	case inputModeGenerateEditConfirm:
		return m.handleGenerateEditConfirmInput(msg)
	default:
		return m.handleTextPromptInput(msg)
	}
}

func (m *Model) handleTextPromptInput(msg tea.KeyMsg) tea.Cmd {
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

func (m *Model) handleGenerateEditConfirmInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n", "N", "enter":
		path := m.input.targetPath
		m.input = inputState{}
		m.setStatus("generated %s", path)
		return nil
	case "y", "Y":
		path := m.input.targetPath
		m.input = inputState{}
		m.setStatus("editing %s", path)
		return editEntryCmd(m.service, path)
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

	switch m.input.mode {
	case inputModeCreateEntry:
		return m.submitCreateEntryPath()
	case inputModeGenerateWizard:
		return m.submitGenerateWizardText()
	default:
		m.input = inputState{}
		return nil
	}
}

func (m *Model) submitCreateEntryPath() tea.Cmd {
	entryPath := strings.Trim(strings.TrimSpace(m.input.value), "/")
	if entryPath == "" {
		m.setStatus("entry path is required")
		return nil
	}

	m.beginGenerateFlow(entryPath, true)
	return nil
}
