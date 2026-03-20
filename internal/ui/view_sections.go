package ui

import (
	"fmt"
	"strings"

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
		if m.input.mode == inputModeDeleteEntries {
			return m.input.prompt
		}

		return m.input.prompt + ": " + m.input.value + "_"
	}

	return "j/k nav • enter open • / search • e edit • n new entry • d delete • space select • x cut • v paste • c copy • p reveal • tab expand • q quit"
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
