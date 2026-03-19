package ui

import (
	"fmt"
	"strings"

	"gopass-tui/internal/tree"
)

// View renders the full TUI.
func (m Model) View() string {
	if len(m.visible) == 0 {
		return "Empty store. Press q to quit."
	}

	var builder strings.Builder

	title := styleTitle.Render("  gopass")
	selectedLabel := ""
	if len(m.selected) > 0 {
		selectedLabel = styleCursor.Render(fmt.Sprintf(" [%d selected]", len(m.selected)))
	}
	cutLabel := ""
	if len(m.cut) > 0 {
		cutLabel = styleCut.Render(fmt.Sprintf(" [%d cut]", len(m.cut)))
	}

	builder.WriteString(title + selectedLabel + cutLabel + "\n")
	builder.WriteString(strings.Repeat("─", min(m.width, 60)) + "\n")
	if m.status != "" {
		builder.WriteString(styleStatus.Render(m.status) + "\n")
	}

	treeHeight := m.height - 5
	if m.status != "" {
		treeHeight--
	}
	if m.preview != "" {
		treeHeight -= min(strings.Count(m.preview, "\n")+2, 8)
	}
	treeHeight = max(treeHeight, 5)

	start := 0
	if m.cursor >= treeHeight {
		start = m.cursor - treeHeight + 1
	}
	end := min(start+treeHeight, len(m.visible))

	for index := start; index < end; index++ {
		visibleNode := m.visible[index]
		indent := strings.Repeat("  ", visibleNode.Depth)
		line := indent + nodePrefix(visibleNode.Node) + visibleNode.Node.Name

		switch {
		case index == m.cursor && m.cut[visibleNode.Node.Path]:
			builder.WriteString(styleCursor.Render("▌") + styleCut.Render(line))
		case index == m.cursor && m.selected[visibleNode.Node.Path]:
			builder.WriteString(styleCursor.Render("▌") + styleSelected.Render(line))
		case index == m.cursor:
			if visibleNode.Node.IsDir {
				builder.WriteString(styleCursor.Render("▌") + styleDir.Render(line))
			} else {
				builder.WriteString(styleCursor.Render("▌" + line))
			}
		case m.selected[visibleNode.Node.Path]:
			builder.WriteString(" " + styleSelected.Render("● "+indent+visibleNode.Node.Name))
		case m.cut[visibleNode.Node.Path]:
			builder.WriteString(" " + styleCut.Render("✂ "+indent+visibleNode.Node.Name))
		case visibleNode.Node.IsDir:
			builder.WriteString(" " + styleDir.Render(line))
		default:
			builder.WriteString(" " + styleEntry.Render(line))
		}

		builder.WriteString("\n")
	}

	if m.preview != "" {
		builder.WriteString("\n")
		previewLines := strings.Split(m.preview, "\n")
		maxLines := min(len(previewLines), 6)
		for _, line := range previewLines[:maxLines] {
			builder.WriteString(stylePreview.Render(line) + "\n")
		}
		if len(previewLines) > maxLines {
			builder.WriteString(stylePreview.Render(fmt.Sprintf("  ... +%d lines", len(previewLines)-maxLines)) + "\n")
		}
	}

	help := "j/k nav • enter open • space select • x cut • v paste • n new dir • c copy • p reveal • tab expand • q quit"
	if m.input.mode != inputModeNone {
		help = m.input.prompt + ": " + m.input.value + "_"
	}
	builder.WriteString("\n" + styleHelp.Render(help))

	return builder.String()
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

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
