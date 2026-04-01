package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View renders the full TUI.
func (m Model) View() string {
	width := max(m.width, defaultViewportWidth)
	height := max(m.height, defaultViewportHeight)
	statusHeight := minStatusPanelHeight
	mainHeight := max(minMainPanelHeight, height-statusHeight-1)
	leftWidth := max(30, width/2)
	rightWidth := max(30, width-leftWidth-1)

	explorerPanel := m.renderExplorerPanel(leftWidth, mainHeight)
	previewPanel := m.renderPreviewPanel(rightWidth, mainHeight)
	mainPanels := lipgloss.JoinHorizontal(lipgloss.Top, explorerPanel, previewPanel)
	statusPanel := m.renderStatusPanel(width, statusHeight)

	base := styleApp.Render(lipgloss.JoinVertical(lipgloss.Left, mainPanels, statusPanel))
	if !m.showHelp {
		return base
	}

	modal := m.renderHelpPanel()
	row := max(1, (height-lipgloss.Height(modal))/2+1)
	col := max(1, (width-lipgloss.Width(modal))/2+1)

	return base + fmt.Sprintf("\x1b[%d;%dH%s\x1b[%d;1H", row, col, modal, height)
}
