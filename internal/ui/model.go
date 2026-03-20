package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
	"gopass-tui/internal/tree"
)

// Model stores the application state for the TUI.
type Model struct {
	service        gopass.Service
	root           *tree.Node
	visible        []tree.FlatNode
	cursor         int
	selected       map[string]bool
	cut            map[string]bool
	preview        string
	previewID      int
	searchQuery    string
	searchExpanded map[string]bool
	searchCursor   int
	status         string
	showPass       bool
	width          int
	height         int
	input          inputState
}

type inputMode int

const (
	inputModeNone inputMode = iota
	inputModeCreateEntry
	inputModeDeleteEntries
	inputModeSearch
)

type inputState struct {
	mode   inputMode
	prompt string
	value  string
	paths  []string
}

type previewLoadedMsg struct {
	requestID int
	path      string
	preview   string
	showPass  bool
	err       error
}

type copyCompletedMsg struct {
	path string
	err  error
}

type editCompletedMsg struct {
	path string
	err  error
}

type createEntryCompletedMsg struct {
	path string
	err  error
}

type deleteCompletedMsg struct {
	focusPath  string
	status     string
	expanded   map[string]bool
	clearPaths []string
}

type treeUpdatedMsg struct {
	root       *tree.Node
	focusPath  string
	status     string
	cut        map[string]bool
	replaceCut bool
	err        error
}

// NewModel builds the initial UI state from the current gopass store.
func NewModel(service gopass.Service) (Model, error) {
	paths, err := service.List(context.Background())
	if err != nil {
		return Model{}, err
	}

	model := Model{
		service:  service,
		root:     tree.Build(paths),
		selected: make(map[string]bool),
		cut:      make(map[string]bool),
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

	case previewLoadedMsg:
		if msg.requestID != m.previewID {
			return m, nil
		}

		node := m.currentNode()
		if node == nil || node.Path != msg.path {
			return m, nil
		}

		if msg.err != nil {
			m.setStatus("error: %v", msg.err)
			m.preview = fmt.Sprintf("(error: %v)", msg.err)
			m.showPass = false
			return m, nil
		}

		m.preview = msg.preview
		m.showPass = msg.showPass

	case copyCompletedMsg:
		if msg.err != nil {
			m.setStatus("error: %v", msg.err)
			return m, nil
		}

		m.setStatus("copied %s to clipboard", msg.path)

	case editCompletedMsg:
		if msg.err != nil {
			m.setStatus("error: %v", msg.err)
			return m, nil
		}

		expanded := m.expandedStateForReload()
		if parent := parentDirectory(msg.path); parent != "" {
			expanded[parent] = true
		}

		m.previewID++
		return m, tea.Batch(
			reloadTreeCmd(m.service, msg.path, fmt.Sprintf("finished editing %s", msg.path), expanded),
			loadPreviewCmd(m.service, m.previewID, msg.path, m.showPass),
		)

	case createEntryCompletedMsg:
		if msg.err != nil {
			m.setStatus("error: %v", msg.err)
			return m, nil
		}

		expanded := m.expandedStateForReload()
		if parent := parentDirectory(msg.path); parent != "" {
			expanded[parent] = true
		}

		m.previewID++
		return m, tea.Batch(
			reloadTreeCmd(m.service, msg.path, fmt.Sprintf("saved %s", msg.path), expanded),
			loadPreviewCmd(m.service, m.previewID, msg.path, false),
		)

	case deleteCompletedMsg:
		for _, path := range msg.clearPaths {
			delete(m.selected, path)
			delete(m.cut, path)
		}
		m.clearPreviewState()
		return m, reloadTreeCmd(m.service, msg.focusPath, msg.status, msg.expanded)

	case treeUpdatedMsg:
		if msg.err != nil {
			m.setStatus("error: %v", msg.err)
			return m, nil
		}

		m.root = msg.root
		m.refresh()
		if msg.focusPath != "" {
			m.focusPath(msg.focusPath)
		}
		if msg.replaceCut {
			m.cut = msg.cut
		}
		if msg.status != "" {
			m.setStatus("%s", msg.status)
		}

	case tea.KeyMsg:
		if m.input.mode != inputModeNone {
			return m, m.handleInput(msg)
		}

		var cmd tea.Cmd

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
			m.clearPreviewState()

		case "G":
			m.cursor = max(0, len(m.visible)-1)
			m.clearPreviewState()

		case "enter", "l", "right":
			cmd = m.handleOpen()

		case "h", "left":
			m.handleBack()

		case "p":
			cmd = m.togglePasswordVisibility()

		case " ":
			m.toggleSelection()

		case "c":
			cmd = m.copyCurrentEntry()

		case "e":
			cmd = m.editCurrentEntry()

		case "x":
			m.cutSelection()

		case "v":
			cmd = m.pasteCutEntries()

		case "n":
			m.beginCreateEntry()

		case "/":
			m.beginSearch()

		case "d":
			m.beginDeleteEntries()

		case "tab":
			m.toggleAllDirectories()
		}

		return m, cmd
	}

	return m, nil
}

func (m *Model) refresh() {
	m.visible = tree.Flatten(m.root, 0)
	m.applySearchFilter()
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

func (m *Model) setStatus(format string, args ...any) {
	m.status = fmt.Sprintf(format, args...)
}

func (m *Model) beginSearch() {
	if m.input.mode != inputModeSearch && m.searchQuery == "" {
		m.searchExpanded = collectExpandedState(m.root)
		m.searchCursor = m.cursor
		m.setAllDirectoriesExpanded(true)
	}

	m.input = inputState{
		mode:   inputModeSearch,
		prompt: "Search",
		value:  m.searchQuery,
	}
	m.applySearchFilter()
}

func (m *Model) applySearchFilter() {
	query := strings.ToLower(strings.TrimSpace(m.searchQuery))
	m.visible = tree.Flatten(m.root, 0)
	if query == "" {
		if m.cursor >= len(m.visible) {
			m.cursor = max(0, len(m.visible)-1)
		}
		return
	}

	filtered := make([]tree.FlatNode, 0, len(m.visible))
	for _, visibleNode := range m.visible {
		if strings.Contains(strings.ToLower(visibleNode.Node.Path), query) {
			filtered = append(filtered, visibleNode)
		}
	}

	m.visible = filtered
	m.cursor = 0
}

func (m *Model) finishSearch(clear bool) {
	if clear {
		m.searchQuery = ""
	}

	m.input = inputState{}
	if clear && m.searchExpanded != nil {
		applyExpandedState(m.root, m.searchExpanded)
		m.searchExpanded = nil
	}
	m.refresh()
	if clear {
		m.cursor = min(m.searchCursor, max(0, len(m.visible)-1))
	}
}

func (m *Model) finishSearchWithSelection() {
	node := m.currentNode()
	focusPath := ""
	if node != nil {
		focusPath = node.Path
	}

	m.searchQuery = ""
	m.input = inputState{}
	m.setAllDirectoriesExpanded(false)
	if focusPath != "" {
		expandPath(m.root, focusPath)
	}
	m.searchExpanded = nil
	m.refresh()
	if focusPath != "" {
		m.focusPath(focusPath)
	}
	if m.currentNode() != nil && m.currentNode().Path == focusPath {
		m.clearPreviewState()
	}
}

func (m *Model) setAllDirectoriesExpanded(expanded bool) {
	setExpandedRecursive(m.root, expanded)
	if m.root != nil {
		m.root.Expanded = true
	}
}

func (m *Model) handleOpen() tea.Cmd {
	node := m.currentNode()
	if node == nil {
		return nil
	}

	if node.IsDir {
		node.Expanded = !node.Expanded
		m.refresh()
		return nil
	}

	m.previewID++
	return loadPreviewCmd(m.service, m.previewID, node.Path, false)
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
			m.clearPreviewState()
			return
		}
	}
}

func (m *Model) focusPath(targetPath string) {
	for index, visibleNode := range m.visible {
		if visibleNode.Node.Path == targetPath {
			m.cursor = index
			return
		}
	}
}
