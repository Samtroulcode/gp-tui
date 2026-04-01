package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"gopass-tui/internal/tree"
)

const (
	baseTreeHeight   = 5
	maxPreviewLines  = 6
	previewLineSlack = 2
	maxPreviewHeight = 8
)

func (m Model) renderHeader() string {
	title := styleTitle.Render("  gopass")
	selectedLabel := ""
	if len(m.selected) > 0 {
		selectedLabel = styleCursor.Render(fmt.Sprintf(" [%d selected]", len(m.selected)))
	}

	cutLabel := ""
	if len(m.cut) > 0 {
		cutLabel = styleCut.Render(fmt.Sprintf(" [%d cut]", len(m.cut)))
	}

	return title + selectedLabel + cutLabel + "\n"
}

func (m Model) renderVisibleNodes() string {
	var builder strings.Builder

	start, end := m.visibleRange()
	for index := start; index < end; index++ {
		builder.WriteString(m.renderVisibleNode(index))
		builder.WriteString("\n")
	}

	return builder.String()
}

func (m Model) visibleRange() (int, int) {
	treeHeight := m.height - baseTreeHeight
	if m.status != "" {
		treeHeight--
	}
	if m.preview != "" {
		treeHeight -= min(strings.Count(m.preview, "\n")+previewLineSlack, maxPreviewHeight)
	}
	if m.showHelp {
		treeHeight -= lipgloss.Height(m.renderHelpPanel())
	}
	treeHeight = max(treeHeight, baseTreeHeight)

	start := 0
	if m.cursor >= treeHeight {
		start = m.cursor - treeHeight + 1
	}

	return start, min(start+treeHeight, len(m.visible))
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
		return " " + styleSelected.Render("● "+indent+node.Name)
	case m.cut[node.Path]:
		return " " + styleCut.Render("✂ "+indent+node.Name)
	case node.IsDir:
		return " " + styleDir.Render(line)
	default:
		return " " + styleEntry.Render(line)
	}
}

func (m Model) renderPreview() string {
	var builder strings.Builder

	builder.WriteString("\n")
	previewLines := strings.Split(m.preview, "\n")
	maxLines := min(len(previewLines), maxPreviewLines)
	for _, line := range previewLines[:maxLines] {
		builder.WriteString(stylePreview.Render(line) + "\n")
	}
	if len(previewLines) > maxLines {
		builder.WriteString(stylePreview.Render(fmt.Sprintf("  ... +%d lines", len(previewLines)-maxLines)) + "\n")
	}

	return builder.String()
}

func (m Model) helpText() string {
	if m.input.mode != inputModeNone {
		if m.input.mode == inputModeDeleteEntries || m.input.promptKind == inputPromptConfirm {
			return m.input.prompt
		}

		return m.input.prompt + ": " + m.input.value + "_"
	}

	if m.showHelp {
		return "? hide help • n new • r regen • R rename • / search • q quit"
	}

	return "? help • / search • n new • r regen • R rename • q quit"
}

func (m Model) renderHelpPanel() string {
	lines := []string{
		"Navigation",
		"  j/k or ↑/↓    Move cursor",
		"  g / G         Jump to top / bottom",
		"  enter or l    Open entry / expand directory",
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
		"Selection & Tree",
		"  space         Select current entry",
		"  x / v         Cut / paste entries",
		"  tab           Toggle all directories",
		"  /             Search full paths",
		"  ?             Toggle help",
		"",
		"Prompts",
		"  enter         Confirm current prompt",
		"  esc           Cancel current prompt",
		"  y / n         Answer confirmation prompts",
		"",
		"Quit",
		"  q or ctrl+c   Quit gp-tui",
	}

	return "\n" + styleHelpPanel.Render(strings.Join(lines, "\n")) + "\n"
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
