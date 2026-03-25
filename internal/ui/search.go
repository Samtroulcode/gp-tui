package ui

import (
	"strings"

	"gopass-tui/internal/tree"
)

func (m *Model) beginSearch() {
	if m.input.mode != inputModeSearch && m.searchQuery == "" {
		m.searchExpanded = collectExpandedState(m.root)
		m.searchCursor = m.cursor
		m.setAllDirectoriesExpanded(true)
	}

	m.input = inputState{
		mode:   inputModeSearch,
		prompt: "Search",
		value:  m.searchQuery,
	}
	m.applySearchFilter()
}

func (m *Model) applySearchFilter() {
	query := strings.ToLower(strings.TrimSpace(m.searchQuery))
	m.visible = tree.Flatten(m.root, 0)
	if query == "" {
		if m.cursor >= len(m.visible) {
			m.cursor = max(0, len(m.visible)-1)
		}
		return
	}

	filtered := make([]tree.FlatNode, 0, len(m.visible))
	for _, visibleNode := range m.visible {
		if strings.Contains(strings.ToLower(visibleNode.Node.Path), query) {
			filtered = append(filtered, visibleNode)
		}
	}

	m.visible = filtered
	m.cursor = 0
}

func (m *Model) finishSearch(clear bool) {
	if clear {
		m.searchQuery = ""
	}

	m.input = inputState{}
	if clear && m.searchExpanded != nil {
		applyExpandedState(m.root, m.searchExpanded)
		m.searchExpanded = nil
	}
	m.refresh()
	if clear {
		m.cursor = min(m.searchCursor, max(0, len(m.visible)-1))
	}
}

func (m *Model) finishSearchWithSelection() {
	node := m.currentNode()
	focusPath := ""
	if node != nil {
		focusPath = node.Path
	}

	m.searchQuery = ""
	m.input = inputState{}
	m.setAllDirectoriesExpanded(false)
	if focusPath != "" {
		expandPath(m.root, focusPath)
	}
	m.searchExpanded = nil
	m.refresh()
	if focusPath != "" {
		m.focusPath(focusPath)
	}
	if m.currentNode() != nil && m.currentNode().Path == focusPath {
		m.clearPreviewState()
	}
}

func (m *Model) setAllDirectoriesExpanded(expanded bool) {
	setExpandedRecursive(m.root, expanded)
	if m.root != nil {
		m.root.Expanded = true
	}
}
