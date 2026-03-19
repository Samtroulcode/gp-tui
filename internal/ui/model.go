package ui

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
	"gopass-tui/internal/tree"
)

// Model stores the application state for the TUI.
type Model struct {
	service   gopass.Service
	root      *tree.Node
	visible   []tree.FlatNode
	cursor    int
	selected  map[string]bool
	cut       map[string]bool
	preview   string
	previewID int
	status    string
	showPass  bool
	width     int
	height    int
	input     inputState
}

type inputMode int

const (
	inputModeNone inputMode = iota
	inputModeCreateDirectory
)

type inputState struct {
	mode   inputMode
	prompt string
	value  string
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
			m.preview = ""

		case "G":
			m.cursor = max(0, len(m.visible)-1)
			m.preview = ""

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

		case "x":
			m.cutSelection()

		case "v":
			cmd = m.pasteCutEntries()

		case "n":
			m.beginCreateDirectory()

		case "tab":
			m.toggleAllDirectories()
		}

		return m, cmd
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

func (m *Model) setStatus(format string, args ...any) {
	m.status = fmt.Sprintf(format, args...)
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
			m.preview = ""
			return
		}
	}
}

func (m *Model) togglePasswordVisibility() tea.Cmd {
	if m.preview == "" {
		return nil
	}

	node := m.currentNode()
	if node == nil || node.IsDir {
		return nil
	}

	m.previewID++
	return loadPreviewCmd(m.service, m.previewID, node.Path, !m.showPass)
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

	delete(m.cut, node.Path)

	if m.cursor < len(m.visible)-1 {
		m.cursor++
	}
}

func (m *Model) copyCurrentEntry() tea.Cmd {
	node := m.currentNode()
	if node == nil || node.IsDir {
		return nil
	}

	return copyEntryCmd(m.service, node.Path)
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

func (m *Model) handleInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.input = inputState{}
		m.setStatus("cancelled")
		return nil
	case "enter":
		return m.submitInput()
	case "backspace":
		if len(m.input.value) == 0 {
			return nil
		}

		runes := []rune(m.input.value)
		m.input.value = string(runes[:len(runes)-1])
		return nil
	}

	if len(msg.Runes) > 0 {
		m.input.value += string(msg.Runes)
	}

	return nil
}

func (m *Model) submitInput() tea.Cmd {
	value := strings.Trim(strings.TrimSpace(m.input.value), "/")
	if value == "" {
		m.setStatus("folder path is required")
		return nil
	}

	switch m.input.mode {
	case inputModeCreateDirectory:
		expanded := m.expandedDirectories()
		parent := parentDirectory(value)
		if parent != "" {
			expanded[parent] = true
		}

		m.input = inputState{}
		m.setStatus("creating folder %s", value)
		return createDirectoryCmd(m.service, value, expanded)
	}

	m.input = inputState{}
	return nil
}

func (m *Model) beginCreateDirectory() {
	base := m.currentDirectory()
	value := ""
	if base != "" {
		value = base + "/"
	}

	m.input = inputState{
		mode:   inputModeCreateDirectory,
		prompt: "New folder",
		value:  value,
	}
}

func (m *Model) cutSelection() {
	paths := m.selectedPaths()
	if len(paths) == 0 {
		node := m.currentNode()
		if node == nil || node.IsDir {
			m.setStatus("select at least one entry to cut")
			return
		}

		paths = []string{node.Path}
	}

	m.cut = make(map[string]bool, len(paths))
	for _, selectedPath := range paths {
		m.cut[selectedPath] = true
	}

	m.selected = make(map[string]bool)
	m.setStatus("cut %s", entryCountLabel(len(paths)))
}

func (m *Model) pasteCutEntries() tea.Cmd {
	paths := m.cutPaths()
	if len(paths) == 0 {
		m.setStatus("cut buffer is empty")
		return nil
	}

	targetDir := m.currentDirectory()
	expanded := m.expandedDirectories()
	if targetDir != "" {
		expanded[targetDir] = true
	}

	m.setStatus("moving %s to %s", entryCountLabel(len(paths)), displayDirectory(targetDir))
	return pasteCutEntriesCmd(m.service, paths, targetDir, expanded)
}

func (m Model) currentDirectory() string {
	node := m.currentNode()
	if node == nil {
		return ""
	}

	if node.IsDir {
		return node.Path
	}

	return parentDirectory(node.Path)
}

func (m Model) selectedPaths() []string {
	paths := make([]string, 0, len(m.selected))
	for selectedPath := range m.selected {
		paths = append(paths, selectedPath)
	}

	sort.Strings(paths)
	return paths
}

func (m Model) cutPaths() []string {
	paths := make([]string, 0, len(m.cut))
	for cutPath := range m.cut {
		paths = append(paths, cutPath)
	}

	sort.Strings(paths)
	return paths
}

func (m Model) expandedDirectories() map[string]bool {
	expanded := make(map[string]bool)
	for _, visibleNode := range m.visible {
		if visibleNode.Node.IsDir && visibleNode.Node.Expanded {
			expanded[visibleNode.Node.Path] = true
		}
	}

	return expanded
}

func (m *Model) focusPath(targetPath string) {
	for index, visibleNode := range m.visible {
		if visibleNode.Node.Path == targetPath {
			m.cursor = index
			return
		}
	}
}

func parentDirectory(entryPath string) string {
	lastSlash := strings.LastIndex(entryPath, "/")
	if lastSlash == -1 {
		return ""
	}

	return entryPath[:lastSlash]
}

func joinPath(directoryPath, name string) string {
	if directoryPath == "" {
		return name
	}

	return directoryPath + "/" + name
}

func displayDirectory(path string) string {
	if path == "" {
		return "store root"
	}

	return path
}

func entryCountLabel(count int) string {
	if count == 1 {
		return "1 entry"
	}

	return fmt.Sprintf("%d entries", count)
}

func applyExpandedState(node *tree.Node, expanded map[string]bool) {
	for _, child := range node.Children {
		if child.IsDir {
			child.Expanded = expanded[child.Path]
			applyExpandedState(child, expanded)
		}
	}
}

func loadPreviewCmd(service gopass.Service, requestID int, entryPath string, showPass bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			preview string
			err     error
		)

		if showPass {
			preview, err = service.Show(ctx, entryPath)
		} else {
			preview, err = service.ShowMasked(ctx, entryPath)
		}

		return previewLoadedMsg{
			requestID: requestID,
			path:      entryPath,
			preview:   preview,
			showPass:  showPass,
			err:       err,
		}
	}
}

func copyEntryCmd(service gopass.Service, entryPath string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		return copyCompletedMsg{path: entryPath, err: service.Copy(ctx, entryPath)}
	}
}

func createDirectoryCmd(service gopass.Service, directoryPath string, expanded map[string]bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := service.Mkdir(ctx, directoryPath); err != nil {
			return treeUpdatedMsg{err: err}
		}

		root, err := loadTree(service, ctx, expanded)
		if err != nil {
			return treeUpdatedMsg{err: err}
		}

		return treeUpdatedMsg{root: root, focusPath: directoryPath, status: fmt.Sprintf("created folder %s", directoryPath)}
	}
}

func pasteCutEntriesCmd(service gopass.Service, paths []string, targetDir string, expanded map[string]bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		failed := make(map[string]bool)
		var (
			moved    int
			firstErr error
		)

		for _, sourcePath := range paths {
			destinationPath := joinPath(targetDir, path.Base(sourcePath))
			if destinationPath == sourcePath {
				continue
			}

			if err := service.Move(ctx, sourcePath, destinationPath); err != nil {
				failed[sourcePath] = true
				if firstErr == nil {
					firstErr = err
				}
				continue
			}

			moved++
		}

		root, err := loadTree(service, ctx, expanded)
		if err != nil {
			return treeUpdatedMsg{err: err}
		}

		status := fmt.Sprintf("moved %s to %s", entryCountLabel(moved), displayDirectory(targetDir))
		if firstErr != nil {
			status = fmt.Sprintf("moved %s, %d failed: %v", entryCountLabel(moved), len(failed), firstErr)
		}

		return treeUpdatedMsg{
			root:       root,
			focusPath:  targetDir,
			status:     status,
			cut:        failed,
			replaceCut: true,
		}
	}
}

func loadTree(service gopass.Service, ctx context.Context, expanded map[string]bool) (*tree.Node, error) {
	paths, err := service.List(ctx)
	if err != nil {
		return nil, err
	}

	root := tree.Build(paths)
	applyExpandedState(root, expanded)
	return root, nil
}
