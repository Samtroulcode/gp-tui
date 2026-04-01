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
	previewPath    string
	previewID      int
	searchQuery    string
	searchExpanded map[string]bool
	searchCursor   int
	status         string
	statusHistory  []string
	showHelp       bool
	showPass       bool
	width          int
	height         int
	input          inputState
}

type inputMode int

const (
	inputModeNone inputMode = iota
	inputModeCreateEntry
	inputModeGenerateWizard
	inputModeGenerateEditConfirm
	inputModeDeleteEntries
	inputModeSearch
	inputModeRenameEntry
)

type inputPromptKind int

const (
	inputPromptText inputPromptKind = iota
	inputPromptConfirm
)

type inputState struct {
	mode        inputMode
	prompt      string
	value       string
	paths       []string
	sourcePath  string
	sourceIsDir bool
	targetPath  string
	promptKind  inputPromptKind
	generation  *generationFlow
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

type generateEntryCompletedMsg struct {
	path        string
	creatingNew bool
	err         error
}

type deleteCompletedMsg struct {
	focusPath  string
	status     string
	expanded   map[string]bool
	clearPaths []string
}

type renameCompletedMsg struct {
	sourcePath       string
	destinationPath  string
	sourceIsDir      bool
	expanded         map[string]bool
	preserveSelected bool
	preserveCut      bool
	err              error
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
		m.previewPath = msg.path
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

	case generateEntryCompletedMsg:
		if msg.err != nil {
			m.setStatus("error: %v", msg.err)
			return m, nil
		}

		if msg.creatingNew {
			m.setStatus("editing %s", msg.path)
			return m, editEntryCmd(m.service, msg.path)
		}

		expanded := m.expandedStateForReload()
		if parent := parentDirectory(msg.path); parent != "" {
			expanded[parent] = true
		}

		m.input = inputState{
			mode:       inputModeGenerateEditConfirm,
			prompt:     fmt.Sprintf("Edit %s now? [y/N]", msg.path),
			targetPath: msg.path,
			promptKind: inputPromptConfirm,
		}
		m.previewID++
		return m, tea.Batch(
			reloadTreeCmd(m.service, msg.path, fmt.Sprintf("generated %s", msg.path), expanded),
			loadPreviewCmd(m.service, m.previewID, msg.path, false),
		)

	case deleteCompletedMsg:
		for _, path := range msg.clearPaths {
			delete(m.selected, path)
			delete(m.cut, path)
		}
		m.clearPreviewState()
		return m, reloadTreeCmd(m.service, msg.focusPath, msg.status, msg.expanded)

	case renameCompletedMsg:
		if msg.err != nil {
			m.setStatus("error: %v", msg.err)
			return m, nil
		}

		remapPathSet(m.selected, msg.sourcePath, msg.destinationPath, msg.sourceIsDir)
		remapPathSet(m.cut, msg.sourcePath, msg.destinationPath, msg.sourceIsDir)
		if msg.preserveSelected {
			m.selected[msg.destinationPath] = true
		}
		if msg.preserveCut {
			m.cut[msg.destinationPath] = true
		}

		status := fmt.Sprintf("renamed %s to %s", msg.sourcePath, msg.destinationPath)
		if msg.sourceIsDir {
			return m, reloadTreeCmd(m.service, msg.destinationPath, status, msg.expanded)
		}

		m.previewID++
		return m, tea.Batch(
			reloadTreeCmd(m.service, msg.destinationPath, status, msg.expanded),
			loadPreviewCmd(m.service, m.previewID, msg.destinationPath, false),
		)

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
				cmd = m.ensureMaskedPreviewForCurrentNode()
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				cmd = m.ensureMaskedPreviewForCurrentNode()
			}

		case "g":
			m.cursor = 0
			cmd = m.ensureMaskedPreviewForCurrentNode()

		case "G":
			m.cursor = max(0, len(m.visible)-1)
			cmd = m.ensureMaskedPreviewForCurrentNode()

		case "enter":
			cmd = m.handleEnter()

		case "l", "right":
			cmd = m.handleOpen()

		case "h", "left":
			cmd = m.handleBack()

		case "p":
			cmd = m.togglePasswordVisibility()

		case " ":
			cmd = m.toggleSelection()

		case "c":
			cmd = m.copyCurrentEntry()

		case "e":
			cmd = m.editCurrentEntry()

		case "r":
			m.beginRegenerateEntry()

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

		case "R":
			m.beginRenameEntry()

		case "?":
			m.showHelp = !m.showHelp

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
	m.previewPath = ""
	m.showPass = false
}

func (m *Model) setStatus(format string, args ...any) {
	m.status = fmt.Sprintf(format, args...)
	if strings.TrimSpace(m.status) == "" {
		return
	}

	m.statusHistory = append(m.statusHistory, m.status)
	if len(m.statusHistory) > 6 {
		m.statusHistory = append([]string(nil), m.statusHistory[len(m.statusHistory)-6:]...)
	}
}
