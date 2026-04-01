package ui

import "github.com/charmbracelet/lipgloss"

// View renders the full TUI.
func (m Model) View() string {
	terminalWidth := max(m.width, 1)
	terminalHeight := max(m.height, 1)
	if m.showHelp {
		return styleApp.
			Width(terminalWidth).
			Height(terminalHeight).
			Render(lipgloss.Place(terminalWidth, terminalHeight, lipgloss.Center, lipgloss.Center, m.renderHelpPanel()))
	}

	width := max(terminalWidth, defaultViewportWidth)
	height := max(terminalHeight, defaultViewportHeight)
	statusHeight := minStatusPanelHeight
	mainHeight := max(minMainPanelHeight, height-statusHeight-1)
	leftWidth := max(30, width/2)
	rightWidth := max(30, width-leftWidth-1)

	explorerPanel := m.renderExplorerPanel(leftWidth, mainHeight)
	previewPanel := m.renderPreviewPanel(rightWidth, mainHeight)
	mainPanels := lipgloss.JoinHorizontal(lipgloss.Top, explorerPanel, previewPanel)
	statusPanel := m.renderStatusPanel(width, statusHeight)

	return styleApp.Render(lipgloss.JoinVertical(lipgloss.Left, mainPanels, statusPanel))
}
