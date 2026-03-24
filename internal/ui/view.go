package ui

import (
	"strings"
)

// View renders the full TUI.
func (m Model) View() string {
	var builder strings.Builder

	builder.WriteString(m.renderHeader())
	builder.WriteString(strings.Repeat("─", min(m.width, 60)) + "\n")
	if m.status != "" {
		builder.WriteString(styleStatus.Render(m.status) + "\n")
	}

	if len(m.visible) == 0 {
		if strings.TrimSpace(m.searchQuery) != "" {
			builder.WriteString("No matching entries.\n")
		} else {
			builder.WriteString("Empty store. Create an entry to get started.\n")
		}
	} else {
		builder.WriteString(m.renderVisibleNodes())
	}

	if m.preview != "" {
		builder.WriteString(m.renderPreview())
	}

	if m.showHelp {
		builder.WriteString(m.renderHelpPanel())
	}

	builder.WriteString("\n" + styleHelp.Render(m.helpText()))

	return builder.String()
}
