package ui

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/tree"
)

type fakeService struct {
	createPath     string
	generatePath   string
	generateLength int
	deleted        []string
	deleteErrs     map[string]error
	listPaths      []string
}

func (f fakeService) List(context.Context) ([]string, error) { return f.listPaths, nil }
func (fakeService) ShowCommand(ctx context.Context, path string) *exec.Cmd {
	return exec.CommandContext(ctx, "true")
}
func (fakeService) SyncCommand(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "true")
}
func (fakeService) Show(context.Context, string) (string, error) {
	return "", errors.New("not implemented")
}
func (fakeService) ShowMasked(context.Context, string) (string, error) {
	return "", errors.New("not implemented")
}
func (fakeService) EditCommand(ctx context.Context, path string) *exec.Cmd {
	return exec.CommandContext(ctx, "true")
}
func (f *fakeService) CreateCommand(ctx context.Context, path string) *exec.Cmd {
	f.createPath = path
	return exec.CommandContext(ctx, "true")
}
func (f *fakeService) Generate(_ context.Context, path string, length int) error {
	f.generatePath = path
	f.generateLength = length
	return nil
}
func (fakeService) Copy(context.Context, string) error { return errors.New("not implemented") }
func (f *fakeService) Delete(_ context.Context, path string) error {
	if err := f.deleteErrs[path]; err != nil {
		return err
	}

	f.deleted = append(f.deleted, path)
	return nil
}
func (fakeService) Move(context.Context, string, string) error { return errors.New("not implemented") }

func TestBeginCreateEntryUsesCurrentDirectory(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"team/api/key"})
	root.Children[0].Expanded = true
	root.Children[0].Children[0].Expanded = true

	model := Model{service: &fakeService{}, root: root}
	model.refresh()
	model.focusPath("team/api/key")

	model.beginCreateEntry()

	if model.input.mode != inputModeCreateEntry {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeCreateEntry)
	}
	if model.input.prompt != "New entry" {
		t.Fatalf("input prompt = %q, want %q", model.input.prompt, "New entry")
	}
	if model.input.value != "team/api/" {
		t.Fatalf("input value = %q, want %q", model.input.value, "team/api/")
	}
}

func TestSubmitInputStartsCreateEntry(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	model.input = inputState{mode: inputModeCreateEntry, value: " team/api/new-secret/ "}

	cmd := model.submitInput()

	if cmd != nil {
		t.Fatal("submitInput returned unexpected cmd")
	}
	if model.input.mode != inputModeCreateGenerateConfirm {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeCreateGenerateConfirm)
	}
	if model.input.prompt != "Generate password? [y/N]" {
		t.Fatalf("prompt = %q, want %q", model.input.prompt, "Generate password? [y/N]")
	}
	if model.input.entryPath != "team/api/new-secret" {
		t.Fatalf("entryPath = %q, want %q", model.input.entryPath, "team/api/new-secret")
	}
	if service.createPath != "" {
		t.Fatalf("create path = %q, want empty", service.createPath)
	}
}

func TestCreateEntryConfirmNoStartsManualCreate(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	model.input = inputState{mode: inputModeCreateGenerateConfirm, prompt: "Generate password? [y/N]", entryPath: "team/api/new-secret"}

	cmd := model.handleCreateGenerateConfirmInput(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("handleCreateGenerateConfirmInput returned nil cmd")
	}
	if model.status != "creating entry team/api/new-secret" {
		t.Fatalf("status = %q, want %q", model.status, "creating entry team/api/new-secret")
	}
	if model.input.mode != inputModeNone {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeNone)
	}
	if service.createPath != "team/api/new-secret" {
		t.Fatalf("create path = %q, want %q", service.createPath, "team/api/new-secret")
	}
}

func TestCreateEntryConfirmYesStartsGenerateLengthPrompt(t *testing.T) {
	t.Parallel()

	model := Model{}
	model.input = inputState{mode: inputModeCreateGenerateConfirm, prompt: "Generate password? [y/N]", entryPath: "team/api/new-secret"}

	cmd := model.handleCreateGenerateConfirmInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if cmd != nil {
		t.Fatal("handleCreateGenerateConfirmInput returned unexpected cmd")
	}
	if model.input.mode != inputModeCreateGenerateLength {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeCreateGenerateLength)
	}
	if model.input.prompt != "Password length" {
		t.Fatalf("prompt = %q, want %q", model.input.prompt, "Password length")
	}
	if model.input.value != "24" {
		t.Fatalf("value = %q, want %q", model.input.value, "24")
	}
}

func TestSubmitInputStartsGeneratedEntryCreation(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	model.input = inputState{mode: inputModeCreateGenerateLength, value: " 32 ", entryPath: "team/api/new-secret"}

	cmd := model.submitInput()
	if cmd == nil {
		t.Fatal("submitInput returned nil cmd")
	}
	if model.status != "generating password for team/api/new-secret" {
		t.Fatalf("status = %q, want %q", model.status, "generating password for team/api/new-secret")
	}
	if model.input.mode != inputModeNone {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeNone)
	}

	msg := cmd()
	generatedMsg, ok := msg.(generateEntryCompletedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want generateEntryCompletedMsg", msg)
	}
	if generatedMsg.path != "team/api/new-secret" {
		t.Fatalf("msg path = %q, want %q", generatedMsg.path, "team/api/new-secret")
	}
	if service.generatePath != "team/api/new-secret" || service.generateLength != 32 {
		t.Fatalf("generate = (%q, %d), want (%q, %d)", service.generatePath, service.generateLength, "team/api/new-secret", 32)
	}
}

func TestSubmitInputRejectsInvalidGeneratedPasswordLength(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	model.input = inputState{mode: inputModeCreateGenerateLength, value: "zero", entryPath: "team/api/new-secret"}

	cmd := model.submitInput()

	if cmd != nil {
		t.Fatal("submitInput returned unexpected cmd")
	}
	if model.status != "password length must be a positive number" {
		t.Fatalf("status = %q, want %q", model.status, "password length must be a positive number")
	}
	if service.generatePath != "" {
		t.Fatalf("generate path = %q, want empty", service.generatePath)
	}
}

func TestViewEmptyStoreShowsCreationHelp(t *testing.T) {
	t.Parallel()

	model := Model{root: tree.Build(nil), width: 40}

	view := model.View()

	if !strings.Contains(view, "Empty store. Create an entry to get started.") {
		t.Fatalf("view = %q, want empty-store message", view)
	}
	if !strings.Contains(view, "n new") {
		t.Fatalf("view = %q, want creation help", view)
	}
}

func TestSubmitInputRequiresEntryPath(t *testing.T) {
	t.Parallel()

	model := Model{service: &fakeService{}}
	model.input = inputState{mode: inputModeCreateEntry, value: " / "}

	cmd := model.submitInput()

	if cmd != nil {
		t.Fatal("submitInput returned unexpected cmd")
	}
	if model.status != "entry path is required" {
		t.Fatalf("status = %q, want %q", model.status, "entry path is required")
	}
}

func TestBeginDeleteEntriesUsesSelectedPaths(t *testing.T) {
	t.Parallel()

	model := Model{selected: map[string]bool{"team/api": true, "team/db": true}}

	model.beginDeleteEntries()

	if model.input.mode != inputModeDeleteEntries {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeDeleteEntries)
	}
	if model.input.prompt != "Delete 2 entries? [y/N]" {
		t.Fatalf("prompt = %q, want %q", model.input.prompt, "Delete 2 entries? [y/N]")
	}
}

func TestSubmitInputDeletesSelectedEntries(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service, selected: map[string]bool{}, cut: map[string]bool{}}
	model.input = inputState{mode: inputModeDeleteEntries, paths: []string{"team/api", "team/db"}}

	cmd := model.submitInput()
	if cmd == nil {
		t.Fatal("submitInput returned nil cmd")
	}
	if model.status != "deleting 2 entries" {
		t.Fatalf("status = %q, want %q", model.status, "deleting 2 entries")
	}

	msg := cmd()
	deleteMsg, ok := msg.(deleteCompletedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want deleteCompletedMsg", msg)
	}
	if len(service.deleted) != 2 {
		t.Fatalf("deleted = %v, want 2 entries", service.deleted)
	}
	if deleteMsg.status != "deleted 2 entries" {
		t.Fatalf("delete status = %q, want %q", deleteMsg.status, "deleted 2 entries")
	}
	if model.input.mode != inputModeNone {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeNone)
	}
}

func TestSubmitInputDeleteKeepsSuccessfulEntriesOnPartialFailure(t *testing.T) {
	t.Parallel()

	service := &fakeService{deleteErrs: map[string]error{"team/db": errors.New("boom")}}
	model := Model{service: service, selected: map[string]bool{}, cut: map[string]bool{}}
	model.input = inputState{mode: inputModeDeleteEntries, paths: []string{"team/api", "team/db"}}

	msg := model.submitInput()()
	deleteMsg, ok := msg.(deleteCompletedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want deleteCompletedMsg", msg)
	}
	if len(deleteMsg.clearPaths) != 1 || deleteMsg.clearPaths[0] != "team/api" {
		t.Fatalf("clearPaths = %v, want [team/api]", deleteMsg.clearPaths)
	}
	if deleteMsg.status != "deleted 1 entry, 1 failed: boom" {
		t.Fatalf("delete status = %q, want %q", deleteMsg.status, "deleted 1 entry, 1 failed: boom")
	}
}

func TestSearchFiltersOnFullPath(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"personal/website/toto/titi", "work/api/token"})

	model := Model{root: root}
	model.refresh()
	model.beginSearch()
	if !root.Children[0].Expanded {
		t.Fatal("search should expand directories")
	}
	model.input.value = "titi"
	model.searchQuery = model.input.value
	model.applySearchFilter()

	if len(model.visible) != 1 {
		t.Fatalf("visible len = %d, want 1", len(model.visible))
	}
	if model.visible[0].Node.Path != "personal/website/toto/titi" {
		t.Fatalf("visible path = %q, want %q", model.visible[0].Node.Path, "personal/website/toto/titi")
	}
	if model.visible[0].Depth != 3 {
		t.Fatalf("visible depth = %d, want 3", model.visible[0].Depth)
	}
}

func TestSearchEscRestoresFullList(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"personal/website/toto/titi", "work/api/token"})
	root.Children[0].Expanded = true
	root.Children[0].Children[0].Expanded = true
	root.Children[0].Children[0].Children[0].Expanded = true

	model := Model{root: root}
	model.refresh()
	fullCount := len(model.visible)
	model.beginSearch()
	model.input.value = "titi"
	model.searchQuery = model.input.value
	model.applySearchFilter()

	model.handleSearchInput(tea.KeyMsg{Type: tea.KeyEsc})

	if model.input.mode != inputModeNone {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeNone)
	}
	if root.Children[0].Expanded != true {
		t.Fatal("expected original expanded state to be restored")
	}
	if len(model.visible) != fullCount {
		t.Fatalf("visible len = %d, want %d", len(model.visible), fullCount)
	}
}

func TestSearchEnterRestoresNormalTreeAroundSelection(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"personal/website/toto/titi", "work/api/token"})

	model := Model{root: root}
	model.refresh()
	model.beginSearch()
	model.input.value = "titi"
	model.searchQuery = model.input.value
	model.applySearchFilter()

	model.handleSearchInput(tea.KeyMsg{Type: tea.KeyEnter})

	if model.input.mode != inputModeNone {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeNone)
	}
	if model.searchQuery != "" {
		t.Fatalf("searchQuery = %q, want empty", model.searchQuery)
	}
	if len(model.visible) <= 1 {
		t.Fatalf("visible len = %d, want expanded normal tree", len(model.visible))
	}
	if !root.Children[0].Expanded {
		t.Fatal("expected selected path ancestors to stay expanded")
	}
	if model.currentNode() == nil || model.currentNode().Path != "personal/website/toto/titi" {
		t.Fatalf("current node = %v, want selected path", model.currentNode())
	}
}

func TestSearchClearingRestoresCurrentTreeState(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"personal/website/toto/titi", "work/api/token"})

	model := Model{root: root}
	model.refresh()
	model.beginSearch()
	model.input.value = "titi"
	model.searchQuery = model.input.value
	model.applySearchFilter()
	model.handleSearchInput(tea.KeyMsg{Type: tea.KeyEnter})

	model.beginSearch()
	model.handleSearchInput(tea.KeyMsg{Type: tea.KeyEsc})

	if !root.Children[0].Expanded {
		t.Fatal("expected current tree state to be restored")
	}
}

func TestSearchEnterWithoutMatchRestoresPreviousState(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"personal/website/toto/titi", "work/api/token", "zeta/item"})
	root.Children[2].Expanded = true

	model := Model{root: root}
	model.refresh()
	model.beginSearch()
	model.input.value = "missing"
	model.searchQuery = model.input.value
	model.applySearchFilter()

	model.handleSearchInput(tea.KeyMsg{Type: tea.KeyEnter})

	if root.Children[1].Expanded {
		t.Fatal("expected collapsed branches to stay collapsed")
	}
	if !root.Children[2].Expanded {
		t.Fatal("expected previous tree state to be restored")
	}
	if model.status != "no matching entry selected" {
		t.Fatalf("status = %q, want %q", model.status, "no matching entry selected")
	}
}

func TestViewShowsNoMatchesForActiveSearch(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"personal/website/toto/titi"})
	root.Children[0].Expanded = true
	root.Children[0].Children[0].Expanded = true
	root.Children[0].Children[0].Children[0].Expanded = true

	model := Model{root: root, width: 40, searchQuery: "missing"}
	model.refresh()
	model.applySearchFilter()

	view := model.View()
	if !strings.Contains(view, "No matching entries.") {
		t.Fatalf("view = %q, want no-match message", view)
	}
}

func TestHelpPanelToggleShowsDetailedHelp(t *testing.T) {
	t.Parallel()

	model := Model{showHelp: true, width: 80}

	view := model.View()

	if !strings.Contains(view, "Navigation  j/k move") {
		t.Fatalf("view = %q, want detailed help", view)
	}
	if !strings.Contains(view, "? hide help") {
		t.Fatalf("view = %q, want compact footer hint", view)
	}
}

func TestQuestionMarkTogglesHelpPanel(t *testing.T) {
	t.Parallel()

	model := Model{}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	opened := updated.(Model)
	if !opened.showHelp {
		t.Fatal("expected help panel to open")
	}

	updated, _ = opened.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	closed := updated.(Model)
	if closed.showHelp {
		t.Fatal("expected help panel to close")
	}
}
