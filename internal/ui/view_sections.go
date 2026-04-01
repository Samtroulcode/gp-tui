package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"gopass-tui/internal/tree"
)

const (
	defaultViewportWidth  = 100
	defaultViewportHeight = 28
	minStatusPanelHeight  = 6
	minMainPanelHeight    = 12
	maxPreviewLines       = 14
	searchBoxHeight       = 3
	panelFrameHeight      = 2
	panelBodyPadding      = 2
	statusHistoryLimit    = 3
)

func (m Model) renderExplorerPanel(width, height int) string {
	width = max(width, 24)
	height = max(height, minMainPanelHeight)
	bodyHeight := max(1, height-panelFrameHeight-searchBoxHeight-panelBodyPadding)

	header := m.renderPanelHeader("Search secrets", fmt.Sprintf("%d/%d", min(m.cursor+1, max(1, len(m.visible))), max(1, len(m.visible))), width-4)
	searchBox := m.renderSearchBox(width - 4)
	treeBody := m.renderTreeBody(bodyHeight)

	content := lipgloss.JoinVertical(lipgloss.Left, header, searchBox, treeBody)
	return stylePanel.Width(width).Height(height).Render(content)
}

func (m Model) renderPreviewPanel(width, height int) string {
	width = max(width, 24)
	height = max(height, minMainPanelHeight)
	bodyHeight := max(1, height-panelFrameHeight-panelBodyPadding)

	node := m.currentNode()
	meta := "No selection"
	if node != nil {
		meta = node.Path
	}

	header := m.renderPanelHeader("Preview", meta, width-4)
	body := m.renderPreviewBody(bodyHeight)
	content := lipgloss.JoinVertical(lipgloss.Left, header, body)

	return stylePanel.Width(width).Height(height).Render(content)
}

func (m Model) renderStatusPanel(width, height int) string {
	width = max(width, 40)
	height = max(height, minStatusPanelHeight)

	lines := []string{
		m.renderPanelHeader("Store status", m.storeMetaSummary(width-4), width-4),
		m.renderStatusSummaryLine(),
		m.renderStoreDetailsLine(),
		m.renderHistoryLine(),
		styleHelp.Render(m.helpText()),
	}

	content := strings.Join(lines, "\n")
	return styleStatusPanel.Width(width).Height(height).Render(content)
}

func (m Model) renderPanelHeader(title, meta string, width int) string {
	left := stylePanelTitle.Render(title)
	if strings.TrimSpace(meta) == "" {
		return left
	}

	availableMetaWidth := max(8, width-lipgloss.Width(title)-2)
	meta = truncateLine(meta, availableMetaWidth)
	right := stylePanelMeta.Render(meta)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, stylePanelMeta.Render("  "), right)
}

func (m Model) renderSearchBox(width int) string {
	width = max(width, 16)
	value := strings.TrimSpace(m.searchQuery)
	placeholder := "/ to search secrets"

	if m.input.mode == inputModeSearch {
		value = m.input.value + "_"
	}
	if value == "" {
		value = stylePlaceholder.Render(placeholder)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styleSearchLabel.Render("Search secrets"),
		searchBoxStyle(m.input.mode == inputModeSearch).Width(width).Render(value),
	)

	return content
}

func (m Model) renderTreeBody(height int) string {
	if len(m.visible) == 0 {
		message := "Empty store. Create an entry to get started."
		if strings.TrimSpace(m.searchQuery) != "" || m.input.mode == inputModeSearch {
			message = "No matching entries."
		}

		return lipgloss.NewStyle().
			Foreground(colorMuted).
			Height(height).
			Render(message)
	}

	var lines []string
	start, end := m.visibleRange(height)
	for index := start; index < end; index++ {
		lines = append(lines, m.renderVisibleNode(index))
	}

	body := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Height(height).Render(body)
}

func (m Model) visibleRange(height int) (int, int) {
	height = max(height, 1)
	start := 0
	if m.cursor >= height {
		start = m.cursor - height + 1
	}

	return start, min(start+height, len(m.visible))
}

func (m Model) renderVisibleNode(index int) string {
	visibleNode := m.visible[index]
	node := visibleNode.Node
	indent := strings.Repeat("  ", visibleNode.Depth)
	line := indent + nodePrefix(node) + node.Name

	switch {
	case index == m.cursor && m.cut[node.Path]:
		return styleCursor.Render("▌") + styleCut.Render(line)
	case index == m.cursor && m.selected[node.Path]:
		return styleCursor.Render("▌") + styleSelected.Render(line)
	case index == m.cursor:
		if node.IsDir {
			return styleCursor.Render("▌") + styleDir.Render(line)
		}

		return styleCursor.Render("▌" + line)
	case m.selected[node.Path]:
		return " " + styleSelected.Render("● "+line)
	case m.cut[node.Path]:
		return " " + styleCut.Render("✂ "+line)
	case node.IsDir:
		return " " + styleDir.Render(line)
	default:
		return " " + styleEntry.Render(line)
	}
}

func (m Model) renderPreviewBody(height int) string {
	node := m.currentNode()
	if node == nil {
		return lipgloss.NewStyle().Height(height).Render(stylePlaceholder.Render("Select an entry to inspect its content."))
	}
	if node.IsDir {
		lines := []string{
			stylePlaceholder.Render("Directory selected."),
			styleStatusLabel.Render("Path: ") + styleStatusValue.Render(displayDirectory(node.Path)),
			styleStatusLabel.Render("Children: ") + styleStatusValue.Render(fmt.Sprintf("%d", len(node.Children))),
			stylePlaceholder.Render("Press enter to expand or collapse this branch."),
		}
		return lipgloss.NewStyle().Height(height).Render(strings.Join(lines, "\n"))
	}

	if strings.TrimSpace(m.preview) == "" {
		lines := []string{
			styleStatusLabel.Render("Path: ") + styleStatusValue.Render(node.Path),
			styleStatusLabel.Render("Visibility: ") + styleStatusValue.Render(previewVisibilityLabel(m.showPass)),
			stylePlaceholder.Render("Masked preview loads automatically when this entry is selected."),
			stylePlaceholder.Render("Use p to reveal or hide the secret after loading."),
		}
		return lipgloss.NewStyle().Height(height).Render(strings.Join(lines, "\n"))
	}

	previewLines := strings.Split(m.preview, "\n")
	maxLines := min(len(previewLines), maxPreviewLines, height)
	lines := []string{
		styleStatusLabel.Render("Path: ") + styleStatusValue.Render(node.Path),
		styleStatusLabel.Render("Visibility: ") + styleStatusValue.Render(previewVisibilityLabel(m.showPass)),
		"",
	}
	for _, line := range previewLines[:maxLines] {
		lines = append(lines, stylePreview.Render(line))
	}
	if len(previewLines) > maxLines {
		lines = append(lines, stylePlaceholder.Render(fmt.Sprintf("… +%d more lines", len(previewLines)-maxLines)))
	}

	return lipgloss.NewStyle().Height(height).Render(strings.Join(lines, "\n"))
}

func (m Model) renderStatusSummaryLine() string {
	parts := []string{
		statusPair("current", m.currentPathLabel()),
		statusPair("entries", fmt.Sprintf("%d visible", len(m.visible))),
		statusPair("selected", fmt.Sprintf("%d", len(m.selected))),
		statusPair("cut", fmt.Sprintf("%d", len(m.cut))),
	}

	if strings.TrimSpace(m.searchQuery) != "" || m.input.mode == inputModeSearch {
		parts = append(parts, statusPair("search", searchableQuery(m)))
	}

	return strings.Join(parts, "  •  ")
}

func (m Model) renderStoreDetailsLine() string {
	parts := []string{
		statusPair("mounts", "pending backend"),
		statusPair("gpg", "pending backend"),
		statusPair("git", "pending backend"),
	}

	if m.showPass {
		parts = append(parts, statusPair("preview", "revealed"))
	} else {
		parts = append(parts, statusPair("preview", "masked"))
	}

	return strings.Join(parts, "  •  ")
}

func (m Model) renderHistoryLine() string {
	history := m.recentStatuses(statusHistoryLimit)
	if len(history) == 0 {
		return statusPair("history", "ready")
	}

	return statusPair("history", strings.Join(history, "  →  "))
}

func (m Model) recentStatuses(limit int) []string {
	if len(m.statusHistory) == 0 {
		return nil
	}
	if limit <= 0 || len(m.statusHistory) <= limit {
		return append([]string(nil), m.statusHistory...)
	}

	return append([]string(nil), m.statusHistory[len(m.statusHistory)-limit:]...)
}

func (m Model) helpText() string {
	if m.input.mode != inputModeNone {
		if m.input.mode == inputModeDeleteEntries || m.input.promptKind == inputPromptConfirm {
			return m.input.prompt
		}
		if m.input.mode == inputModeSearch {
			return "Search is focused • type to filter • enter to keep selection • esc to cancel"
		}

		return m.input.prompt + ": " + m.input.value + "_"
	}

	if m.showHelp {
		return "? hide help • / search • enter edit • n new • R rename • q quit"
	}

	return "? help • / search • enter edit • n new • R rename • q quit"
}

func (m Model) renderHelpPanel() string {
	lines := []string{
		stylePanelTitle.Render("gp-tui help"),
		"",
		"Navigation",
		"  j/k or ↑/↓    Move cursor",
		"  g / G         Jump to top / bottom",
		"  enter         Edit entry / expand directory",
		"  l or right    Expand directory / refresh entry preview",
		"  h             Go back / collapse directory",
		"",
		"Entries",
		"  n             New entry",
		"  r             Regenerate current password",
		"  R             Rename or move current entry",
		"  e             Edit current entry",
		"  c             Copy current entry",
		"  p             Reveal / hide password preview",
		"  d             Delete current or selected entries",
		"",
		"Selection & Search",
		"  space         Select current entry",
		"  x / v         Cut / paste entries",
		"  tab           Toggle all directories",
		"  /             Focus the persistent search field",
		"  ?             Toggle this help modal",
		"",
		"Prompts",
		"  enter         Confirm current prompt",
		"  esc           Cancel current prompt",
		"  y / n         Answer confirmation prompts",
		"",
		"Quit",
		"  q or ctrl+c   Quit gp-tui",
	}

	return styleModal.Width(54).Render(strings.Join(lines, "\n"))
}

func (m Model) currentPathLabel() string {
	node := m.currentNode()
	if node == nil {
		return "none"
	}
	if node.Path == "" {
		return "store root"
	}

	return node.Path
}

func (m Model) storeMetaSummary(width int) string {
	parts := []string{"future-ready", "fzf slot", displayDirectory(m.currentDirectory())}
	return truncateLine(strings.Join(parts, "  •  "), width)
}

func nodePrefix(node *tree.Node) string {
	if !node.IsDir {
		return "  "
	}

	if node.Expanded {
		return "▾ "
	}

	return "▸ "
}

func previewVisibilityLabel(showPass bool) string {
	if showPass {
		return "revealed"
	}

	return "masked"
}

func searchBoxStyle(active bool) lipgloss.Style {
	if active {
		return styleSearchBoxActive
	}

	return styleSearchBox
}

func statusPair(label, value string) string {
	return styleStatusLabel.Render(label+":") + " " + styleStatusValue.Render(value)
}

func searchableQuery(m Model) string {
	if m.input.mode == inputModeSearch && strings.TrimSpace(m.input.value) != "" {
		return m.input.value
	}
	if strings.TrimSpace(m.searchQuery) != "" {
		return m.searchQuery
	}

	return "active"
}

func truncateLine(value string, width int) string {
	if width <= 0 || lipgloss.Width(value) <= width {
		return value
	}
	if width <= 1 {
		return "…"
	}

	runes := []rune(value)
	if len(runes) >= width {
		return string(runes[:width-1]) + "…"
	}

	return value
}
