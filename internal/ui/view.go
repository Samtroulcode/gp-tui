package ui

import (
	"strings"
)

// View renders the full TUI.
func (m Model) View() string {
	if len(m.visible) == 0 {
		return "Empty store. Press q to quit."
	}

	var builder strings.Builder

	builder.WriteString(m.renderHeader())
	builder.WriteString(strings.Repeat("─", min(m.width, 60)) + "\n")
	if m.status != "" {
		builder.WriteString(styleStatus.Render(m.status) + "\n")
	}

	builder.WriteString(m.renderVisibleNodes())

	if m.preview != "" {
		builder.WriteString(m.renderPreview())
	}

	builder.WriteString("\n" + styleHelp.Render(m.helpText()))

	return builder.String()
}
