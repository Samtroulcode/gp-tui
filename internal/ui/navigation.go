package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) handleOpen() tea.Cmd {
	node := m.currentNode()
	if node == nil {
		return nil
	}

	if node.IsDir {
		node.Expanded = !node.Expanded
		m.refresh()
		return nil
	}

	m.previewID++
	return loadPreviewCmd(m.service, m.previewID, node.Path, false)
}

func (m *Model) handleBack() {
	node := m.currentNode()
	if node == nil {
		return
	}

	if node.IsDir && node.Expanded {
		node.Expanded = false
		m.refresh()
		return
	}

	for index := m.cursor - 1; index >= 0; index-- {
		if m.visible[index].Node.IsDir && m.visible[index].Depth < m.visible[m.cursor].Depth {
			m.cursor = index
			m.clearPreviewState()
			return
		}
	}
}

func (m *Model) focusPath(targetPath string) {
	for index, visibleNode := range m.visible {
		if visibleNode.Node.Path == targetPath {
			m.cursor = index
			return
		}
	}
}
