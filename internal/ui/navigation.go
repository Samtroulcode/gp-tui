package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) handleEnter() tea.Cmd {
	node := m.currentNode()
	if node == nil {
		return nil
	}

	if node.IsDir {
		return m.handleOpen()
	}

	return m.editCurrentEntry()
}

func (m *Model) handleOpen() tea.Cmd {
	node := m.currentNode()
	if node == nil {
		return nil
	}

	if node.IsDir {
		node.Expanded = !node.Expanded
		m.refresh()
		return m.ensureMaskedPreviewForCurrentNode()
	}

	return m.ensureMaskedPreviewForCurrentNode()
}

func (m *Model) handleBack() tea.Cmd {
	node := m.currentNode()
	if node == nil {
		return nil
	}

	if node.IsDir && node.Expanded {
		node.Expanded = false
		m.refresh()
		return m.ensureMaskedPreviewForCurrentNode()
	}

	for index := m.cursor - 1; index >= 0; index-- {
		if m.visible[index].Node.IsDir && m.visible[index].Depth < m.visible[m.cursor].Depth {
			m.cursor = index
			return m.ensureMaskedPreviewForCurrentNode()
		}
	}

	return nil
}

func (m *Model) focusPath(targetPath string) {
	for index, visibleNode := range m.visible {
		if visibleNode.Node.Path == targetPath {
			m.cursor = index
			return
		}
	}
}

func (m *Model) ensureMaskedPreviewForCurrentNode() tea.Cmd {
	node := m.currentNode()
	if node == nil || node.IsDir {
		m.clearPreviewState()
		return nil
	}
	if m.previewPath == node.Path && m.preview != "" && !m.showPass {
		return nil
	}

	m.clearPreviewState()
	m.previewID++
	return loadPreviewCmd(m.service, m.previewID, node.Path, false)
}
