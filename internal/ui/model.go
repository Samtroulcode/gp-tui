package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
	"gopass-tui/internal/tree"
)

// Model stores the application state for the TUI.
type Model struct {
	service  gopass.Service
	root     *tree.Node
	visible  []tree.FlatNode
	cursor   int
	selected map[string]bool
	preview  string
	showPass bool
	width    int
	height   int
}

// NewModel builds the initial UI state from the current gopass store.
func NewModel(service gopass.Service) (Model, error) {
	paths, err := service.List()
	if err != nil {
		return Model{}, err
	}

	model := Model{
		service:  service,
		root:     tree.Build(paths),
		selected: make(map[string]bool),
	}
	model.refresh()

	return model, nil
}

// Init satisfies the Bubble Tea model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "j", "down":
			if m.cursor < len(m.visible)-1 {
				m.cursor++
				m.clearPreviewState()
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.clearPreviewState()
			}

		case "g":
			m.cursor = 0
			m.preview = ""

		case "G":
			m.cursor = max(0, len(m.visible)-1)
			m.preview = ""

		case "enter", "l", "right":
			m.handleOpen()

		case "h", "left":
			m.handleBack()

		case "p":
			m.togglePasswordVisibility()

		case " ":
			m.toggleSelection()

		case "c":
			m.copyCurrentEntry()

		case "tab":
			m.toggleAllDirectories()
		}
	}

	return m, nil
}

func (m *Model) refresh() {
	m.visible = tree.Flatten(m.root, 0)
	if m.cursor >= len(m.visible) {
		m.cursor = max(0, len(m.visible)-1)
	}
}

func (m Model) currentNode() *tree.Node {
	if m.cursor >= 0 && m.cursor < len(m.visible) {
		return m.visible[m.cursor].Node
	}

	return nil
}

func (m *Model) clearPreviewState() {
	m.preview = ""
	m.showPass = false
}

func (m *Model) handleOpen() {
	node := m.currentNode()
	if node == nil {
		return
	}

	if node.IsDir {
		node.Expanded = !node.Expanded
		m.refresh()
		return
	}

	preview, err := m.service.ShowMasked(node.Path)
	if err != nil {
		m.preview = fmt.Sprintf("(error: %v)", err)
		m.showPass = false
		return
	}

	m.preview = preview
	m.showPass = false
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
			m.preview = ""
			return
		}
	}
}

func (m *Model) togglePasswordVisibility() {
	if m.preview == "" {
		return
	}

	node := m.currentNode()
	if node == nil || node.IsDir {
		return
	}

	m.showPass = !m.showPass

	var (
		preview string
		err     error
	)

	if m.showPass {
		preview, err = m.service.Show(node.Path)
	} else {
		preview, err = m.service.ShowMasked(node.Path)
	}

	if err != nil {
		m.preview = fmt.Sprintf("(error: %v)", err)
		m.showPass = false
		return
	}

	m.preview = preview
}

func (m *Model) toggleSelection() {
	node := m.currentNode()
	if node == nil || node.IsDir {
		return
	}

	m.selected[node.Path] = !m.selected[node.Path]
	if !m.selected[node.Path] {
		delete(m.selected, node.Path)
	}

	if m.cursor < len(m.visible)-1 {
		m.cursor++
	}
}

func (m *Model) copyCurrentEntry() {
	node := m.currentNode()
	if node == nil || node.IsDir {
		return
	}

	if err := m.service.Copy(node.Path); err != nil {
		m.preview = fmt.Sprintf("(error: %v)", err)
		return
	}

	m.preview = "✓ copied to clipboard"
}

func (m *Model) toggleAllDirectories() {
	allExpanded := true
	for _, visibleNode := range m.visible {
		if visibleNode.Node.IsDir && !visibleNode.Node.Expanded {
			allExpanded = false
			break
		}
	}

	for _, visibleNode := range m.visible {
		if visibleNode.Node.IsDir {
			visibleNode.Node.Expanded = !allExpanded
		}
	}

	m.root.Expanded = true
	m.refresh()
}
