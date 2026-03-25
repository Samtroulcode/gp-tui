package ui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) togglePasswordVisibility() tea.Cmd {
	if m.preview == "" {
		return nil
	}

	node := m.currentNode()
	if node == nil || node.IsDir {
		return nil
	}

	m.previewID++
	return loadPreviewCmd(m.service, m.previewID, node.Path, !m.showPass)
}

func (m *Model) toggleSelection() {
	node := m.currentNode()
	if node == nil || node.IsDir {
		return
	}

	m.selected[node.Path] = !m.selected[node.Path]
	if !m.selected[node.Path] {
		delete(m.selected, node.Path)
	}

	delete(m.cut, node.Path)

	if m.cursor < len(m.visible)-1 {
		m.cursor++
		m.clearPreviewState()
	}
}

func (m *Model) copyCurrentEntry() tea.Cmd {
	node := m.currentNode()
	if node == nil || node.IsDir {
		return nil
	}

	return copyEntryCmd(m.service, node.Path)
}

func (m *Model) editCurrentEntry() tea.Cmd {
	node := m.currentNode()
	if node == nil || node.IsDir {
		return nil
	}

	m.setStatus("editing %s", node.Path)
	return editEntryCmd(m.service, node.Path)
}

func (m *Model) beginRegenerateEntry() {
	node := m.currentNode()
	if node == nil || node.IsDir {
		m.setStatus("select an entry to regenerate")
		return
	}

	m.beginGenerateFlow(node.Path, false)
}

func (m *Model) toggleAllDirectories() {
	allExpanded := true
	for _, visibleNode := range m.visible {
		if visibleNode.Node.IsDir && !visibleNode.Node.Expanded {
			allExpanded = false
			break
		}
	}

	for _, visibleNode := range m.visible {
		if visibleNode.Node.IsDir {
			visibleNode.Node.Expanded = !allExpanded
		}
	}

	m.root.Expanded = true
	m.refresh()
}

func (m *Model) cutSelection() {
	paths := m.selectedPaths()
	if len(paths) == 0 {
		node := m.currentNode()
		if node == nil || node.IsDir {
			m.setStatus("select at least one entry to cut")
			return
		}

		paths = []string{node.Path}
	}

	m.cut = make(map[string]bool, len(paths))
	for _, selectedPath := range paths {
		m.cut[selectedPath] = true
	}

	m.selected = make(map[string]bool)
	m.setStatus("cut %s", entryCountLabel(len(paths)))
}

func (m *Model) pasteCutEntries() tea.Cmd {
	paths := m.cutPaths()
	if len(paths) == 0 {
		m.setStatus("cut buffer is empty")
		return nil
	}

	targetDir := m.currentDirectory()
	expanded := m.expandedStateForReload()
	if targetDir != "" {
		expanded[targetDir] = true
	}

	m.setStatus("moving %s to %s", entryCountLabel(len(paths)), displayDirectory(targetDir))
	return pasteCutEntriesCmd(m.service, paths, targetDir, expanded)
}

func (m *Model) beginDeleteEntries() {
	paths := m.selectedPaths()
	if len(paths) == 0 {
		node := m.currentNode()
		if node == nil || node.IsDir {
			m.setStatus("select at least one entry to delete")
			return
		}

		paths = []string{node.Path}
	}

	m.input = inputState{
		mode:   inputModeDeleteEntries,
		prompt: "Delete " + entryCountLabel(len(paths)) + "? [y/N]",
		paths:  paths,
	}
}

func (m Model) currentDirectory() string {
	node := m.currentNode()
	if node == nil {
		return ""
	}

	if node.IsDir {
		return node.Path
	}

	return parentDirectory(node.Path)
}

func (m Model) selectedPaths() []string {
	paths := make([]string, 0, len(m.selected))
	for selectedPath := range m.selected {
		paths = append(paths, selectedPath)
	}

	sort.Strings(paths)
	return paths
}

func (m Model) cutPaths() []string {
	paths := make([]string, 0, len(m.cut))
	for cutPath := range m.cut {
		paths = append(paths, cutPath)
	}

	sort.Strings(paths)
	return paths
}

func (m Model) expandedDirectories() map[string]bool {
	return collectExpandedState(m.root)
}

func (m Model) expandedStateForReload() map[string]bool {
	if strings.TrimSpace(m.searchQuery) != "" {
		expanded := make(map[string]bool)
		markAllDirectoriesExpanded(m.root, expanded)
		return expanded
	}

	return m.expandedDirectories()
}
