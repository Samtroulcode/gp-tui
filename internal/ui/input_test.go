package ui

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
	"gopass-tui/internal/tree"
)

type fakeService struct {
	createPath      string
	editPath        string
	generateRequest gopass.GenerateRequest
	deleted         []string
	deleteErrs      map[string]error
	movedFrom       string
	movedTo         string
	moveErr         error
	maskedPath      string
	maskedValue     string
	maskedErr       error
	listPaths       []string
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
func (f *fakeService) ShowMasked(_ context.Context, path string) (string, error) {
	if f.maskedErr != nil {
		return "", f.maskedErr
	}

	f.maskedPath = path
	if f.maskedValue == "" {
		return "******", nil
	}

	return f.maskedValue, nil
}
func (f *fakeService) EditCommand(ctx context.Context, path string) *exec.Cmd {
	f.editPath = path
	return exec.CommandContext(ctx, "true")
}
func (f *fakeService) CreateCommand(ctx context.Context, path string) *exec.Cmd {
	f.createPath = path
	return exec.CommandContext(ctx, "true")
}
func (f *fakeService) GenerateCommand(ctx context.Context, request gopass.GenerateRequest) (*exec.Cmd, error) {
	f.generateRequest = request
	return exec.CommandContext(ctx, "true"), nil
}
func (fakeService) Copy(context.Context, string) error { return errors.New("not implemented") }
func (f *fakeService) Delete(_ context.Context, path string) error {
	if err := f.deleteErrs[path]; err != nil {
		return err
	}

	f.deleted = append(f.deleted, path)
	return nil
}
func (f *fakeService) Move(_ context.Context, sourcePath, destinationPath string) error {
	if f.moveErr != nil {
		return f.moveErr
	}

	f.movedFrom = sourcePath
	f.movedTo = destinationPath
	return nil
}

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
	if model.input.mode != inputModeGenerateWizard {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeGenerateWizard)
	}
	if model.input.prompt != "Generate password? [y/N]" {
		t.Fatalf("prompt = %q, want %q", model.input.prompt, "Generate password? [y/N]")
	}
	if model.input.generation == nil || model.input.generation.request.Path != "team/api/new-secret" {
		t.Fatalf("generation path = %v, want %q", model.input.generation, "team/api/new-secret")
	}
	if service.createPath != "" {
		t.Fatalf("create path = %q, want empty", service.createPath)
	}
}

func TestBeginRenameEntryUsesParentDirectoryPrefix(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"ssh/test"})
	root.Children[0].Expanded = true

	model := Model{service: &fakeService{}, root: root}
	model.refresh()
	model.focusPath("ssh/test")

	model.beginRenameEntry()

	if model.input.mode != inputModeRenameEntry {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeRenameEntry)
	}
	if model.input.prompt != "Rename entry" {
		t.Fatalf("input prompt = %q, want %q", model.input.prompt, "Rename entry")
	}
	if model.input.value != "ssh/" {
		t.Fatalf("input value = %q, want %q", model.input.value, "ssh/")
	}
}

func TestSubmitInputRenamesEntryToTypedDestination(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service, selected: map[string]bool{"ssh/test": true}, cut: map[string]bool{}}
	model.input = inputState{mode: inputModeRenameEntry, sourcePath: "ssh/test", value: " ssh/prod ", sourceIsDir: false}

	cmd := model.submitInput()
	if cmd == nil {
		t.Fatal("submitInput returned nil cmd")
	}
	if model.status != "renaming ssh/test" {
		t.Fatalf("status = %q, want %q", model.status, "renaming ssh/test")
	}
	if model.input.mode != inputModeNone {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeNone)
	}

	msg := cmd()
	renameMsg, ok := msg.(renameCompletedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want renameCompletedMsg", msg)
	}
	if service.movedFrom != "ssh/test" || service.movedTo != "ssh/prod" {
		t.Fatalf("move = %q -> %q, want %q -> %q", service.movedFrom, service.movedTo, "ssh/test", "ssh/prod")
	}
	if renameMsg.destinationPath != "ssh/prod" {
		t.Fatalf("destination path = %q, want %q", renameMsg.destinationPath, "ssh/prod")
	}
	if !renameMsg.preserveSelected {
		t.Fatal("expected rename to preserve selection state")
	}
}

func TestSubmitInputRejectsEmptyRenameDestination(t *testing.T) {
	t.Parallel()

	model := Model{service: &fakeService{}}
	model.input = inputState{mode: inputModeRenameEntry, sourcePath: "ssh/test", value: " / "}

	cmd := model.submitInput()

	if cmd != nil {
		t.Fatal("submitInput returned unexpected cmd")
	}
	if model.status != "destination path is required" {
		t.Fatalf("status = %q, want %q", model.status, "destination path is required")
	}
}

func TestRenameDirectoryRemapsNestedSelectionAndCutPaths(t *testing.T) {
	t.Parallel()

	model := Model{
		selected: map[string]bool{"ssh/team/key": true},
		cut:      map[string]bool{"ssh/team/other": true},
	}

	updated, cmd := model.Update(renameCompletedMsg{
		sourcePath:      "ssh/team",
		destinationPath: "infra/team",
		sourceIsDir:     true,
		expanded:        map[string]bool{"infra": true},
	})
	if cmd == nil {
		t.Fatal("Update returned nil cmd")
	}

	model = updated.(Model)
	if !model.selected["infra/team/key"] {
		t.Fatalf("selected = %v, want remapped nested path", model.selected)
	}
	if !model.cut["infra/team/other"] {
		t.Fatalf("cut = %v, want remapped nested path", model.cut)
	}
	if model.selected["ssh/team/key"] || model.cut["ssh/team/other"] {
		t.Fatalf("stale paths left behind: selected=%v cut=%v", model.selected, model.cut)
	}
}

func TestCreateEntryConfirmNoStartsManualCreate(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	model.beginGenerateFlow("team/api/new-secret", true)

	cmd := model.handleGenerateWizardConfirmInput(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("handleGenerateWizardConfirmInput returned nil cmd")
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

func TestCreateEntryConfirmYesStartsQuickGenerationPrompt(t *testing.T) {
	t.Parallel()

	model := Model{}
	model.beginGenerateFlow("team/api/new-secret", true)

	cmd := model.handleGenerateWizardConfirmInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if cmd != nil {
		t.Fatal("handleGenerateWizardConfirmInput returned unexpected cmd")
	}
	if model.input.mode != inputModeGenerateWizard {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeGenerateWizard)
	}
	if model.input.generation == nil || model.input.generation.step != generateStepQuickConfirm {
		t.Fatalf("step = %v, want %v", model.input.generation, generateStepQuickConfirm)
	}
	if model.input.prompt != "Quick generation with recommended defaults? [Y/n]" {
		t.Fatalf("prompt = %q, want %q", model.input.prompt, "Quick generation with recommended defaults? [Y/n]")
	}
	if model.input.promptKind != inputPromptConfirm {
		t.Fatalf("prompt kind = %v, want %v", model.input.promptKind, inputPromptConfirm)
	}
}

func TestQuickGenerationUsesRecommendedDefaults(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	flow := &generationFlow{
		creatingNew: true,
		request: gopass.GenerateRequest{
			Path:      "team/api/new-secret",
			Length:    defaultQuickPasswordLength,
			Generator: "cryptic",
			Language:  "en",
		},
		step: generateStepQuickConfirm,
	}
	model.input = inputState{mode: inputModeGenerateWizard, promptKind: inputPromptConfirm, generation: flow}

	cmd := model.submitGenerateWizardConfirm(true)
	if cmd == nil {
		t.Fatal("submitGenerateWizardConfirm returned nil cmd")
	}
	if model.status != "generating password for team/api/new-secret" {
		t.Fatalf("status = %q, want %q", model.status, "generating password for team/api/new-secret")
	}
	if model.input.mode != inputModeNone {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeNone)
	}
	want := gopass.GenerateRequest{
		Path:      "team/api/new-secret",
		Length:    defaultQuickPasswordLength,
		Generator: "cryptic",
		Language:  "en",
		Symbols:   true,
		Strict:    true,
	}
	if !reflect.DeepEqual(service.generateRequest, want) {
		t.Fatalf("generate request = %#v, want %#v", service.generateRequest, want)
	}
}

func TestQuickGenerationEnterUsesRecommendedDefaults(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	flow := &generationFlow{
		creatingNew: true,
		request:     gopass.GenerateRequest{Path: "team/api/new-secret", Length: defaultQuickPasswordLength, Generator: "cryptic", Language: "en"},
		step:        generateStepQuickConfirm,
	}
	model.input = inputState{mode: inputModeGenerateWizard, promptKind: inputPromptConfirm, generation: flow}

	cmd := model.handleGenerateWizardConfirmInput(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("handleGenerateWizardConfirmInput returned nil cmd")
	}
	if service.generateRequest.Path != "team/api/new-secret" {
		t.Fatalf("generate path = %q, want %q", service.generateRequest.Path, "team/api/new-secret")
	}
}

func TestDecliningQuickGenerationStartsFullWizard(t *testing.T) {
	t.Parallel()

	model := Model{}
	flow := &generationFlow{
		creatingNew: true,
		request:     gopass.GenerateRequest{Path: "team/api/new-secret", Length: defaultQuickPasswordLength, Generator: "cryptic", Language: "en"},
		step:        generateStepQuickConfirm,
	}
	model.input = inputState{mode: inputModeGenerateWizard, promptKind: inputPromptConfirm, generation: flow}

	cmd := model.submitGenerateWizardConfirm(false)

	if cmd != nil {
		t.Fatal("submitGenerateWizardConfirm returned unexpected cmd")
	}
	if model.input.prompt != "Secret key (blank for password line)" {
		t.Fatalf("prompt = %q, want secret key prompt", model.input.prompt)
	}
	if model.input.generation == nil || model.input.generation.step != generateStepKey {
		t.Fatalf("step = %v, want %v", model.input.generation, generateStepKey)
	}
}

func TestSubmitInputRejectsInvalidGeneratedPasswordLength(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	model.input = inputState{
		mode:       inputModeGenerateWizard,
		prompt:     "Password length",
		value:      "zero",
		generation: &generationFlow{request: gopass.GenerateRequest{Path: "team/api/new-secret", Length: 24, Generator: "cryptic", Language: "en"}, step: generateStepLength},
	}

	cmd := model.submitInput()

	if cmd != nil {
		t.Fatal("submitInput returned unexpected cmd")
	}
	if model.status != "password length must be a positive number" {
		t.Fatalf("status = %q, want %q", model.status, "password length must be a positive number")
	}
	if !reflect.DeepEqual(service.generateRequest, gopass.GenerateRequest{}) {
		t.Fatalf("generate request = %#v, want empty", service.generateRequest)
	}
}

func TestFullWizardCrypticFlowCollectsRelevantOptions(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	flow := &generationFlow{creatingNew: true, request: gopass.GenerateRequest{Path: "team/api/new-secret", Length: 24, Generator: "cryptic", Language: "en"}, step: generateStepGenerator}
	model.input = inputState{mode: inputModeGenerateWizard, prompt: "Generator [cryptic|memorable|xkcd|external]", value: "cryptic", generation: flow}

	cmd := model.submitGenerateWizardText()
	if cmd != nil || model.input.generation.step != generateStepLength {
		t.Fatal("expected generator step to move to length")
	}

	model.input.value = "28"
	cmd = model.submitGenerateWizardText()
	if cmd != nil || model.input.generation.step != generateStepSymbols {
		t.Fatal("expected length step to move to symbols")
	}

	cmd = model.submitGenerateWizardConfirm(true)
	if cmd != nil || model.input.generation.step != generateStepStrict {
		t.Fatal("expected symbols step to move to strict")
	}

	cmd = model.submitGenerateWizardConfirm(true)
	if cmd == nil {
		t.Fatal("expected strict confirmation to trigger generation")
	}

	want := gopass.GenerateRequest{Path: "team/api/new-secret", Length: 28, Generator: "cryptic", Language: "en", Symbols: true, Strict: true}
	if !reflect.DeepEqual(service.generateRequest, want) {
		t.Fatalf("generate request = %#v, want %#v", service.generateRequest, want)
	}
}

func TestFullWizardXKCDFlowCollectsSeparatorAndLanguage(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}
	flow := &generationFlow{creatingNew: false, request: gopass.GenerateRequest{Path: "team/api/key", Length: 24, Generator: "cryptic", Language: "en", Force: true}, step: generateStepGenerator}
	model.input = inputState{mode: inputModeGenerateWizard, prompt: "Generator [cryptic|memorable|xkcd|external]", value: "xkcd", generation: flow}

	_ = model.submitGenerateWizardText()
	model.input.value = "5"
	_ = model.submitGenerateWizardText()
	model.input.value = "-"
	_ = model.submitGenerateWizardText()
	model.input.value = "de"
	cmd := model.submitGenerateWizardText()

	if cmd == nil {
		t.Fatal("expected language step to trigger generation")
	}

	want := gopass.GenerateRequest{Path: "team/api/key", Length: 5, Generator: "xkcd", Language: "de", Force: true, Separator: "-"}
	if !reflect.DeepEqual(service.generateRequest, want) {
		t.Fatalf("generate request = %#v, want %#v", service.generateRequest, want)
	}
}

func TestBeginRegenerateEntryStartsOverwriteConfirmation(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"team/api/key"})
	root.Children[0].Expanded = true
	root.Children[0].Children[0].Expanded = true

	model := Model{service: &fakeService{}, root: root}
	model.refresh()
	model.focusPath("team/api/key")

	model.beginRegenerateEntry()

	if model.input.mode != inputModeGenerateWizard {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeGenerateWizard)
	}
	if model.input.promptKind != inputPromptConfirm {
		t.Fatalf("prompt kind = %v, want %v", model.input.promptKind, inputPromptConfirm)
	}
	if !strings.Contains(model.input.prompt, "overwrite the current password") {
		t.Fatalf("prompt = %q, want overwrite warning", model.input.prompt)
	}
}

func TestRegenerateEntryConfirmationAdvancesToSharedGenerateWizard(t *testing.T) {
	t.Parallel()

	model := Model{}
	model.beginGenerateFlow("team/api/key", false)

	cmd := model.handleGenerateWizardConfirmInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if cmd != nil {
		t.Fatal("handleGenerateWizardConfirmInput returned unexpected cmd")
	}
	if model.input.generation == nil || model.input.generation.step != generateStepQuickConfirm {
		t.Fatalf("step = %v, want %v", model.input.generation, generateStepQuickConfirm)
	}
	if !model.input.generation.request.Force {
		t.Fatal("expected regenerate flow to force overwrite after confirmation")
	}
}

func TestGenerateCompletionForNewEntryStartsEditingImmediately(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service}

	updated, cmd := model.Update(generateEntryCompletedMsg{path: "team/api/new-secret", creatingNew: true})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("Update returned nil cmd")
	}
	if model.status != "editing team/api/new-secret" {
		t.Fatalf("status = %q, want editing status", model.status)
	}
	if service.editPath != "team/api/new-secret" {
		t.Fatalf("edit path = %q, want %q", service.editPath, "team/api/new-secret")
	}
}

func TestGenerateCompletionForExistingEntryAsksForEdit(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	model := Model{service: service, root: tree.Build([]string{"team/api/key"})}

	updated, cmd := model.Update(generateEntryCompletedMsg{path: "team/api/key", creatingNew: false})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("Update returned nil cmd")
	}
	if model.input.mode != inputModeGenerateEditConfirm {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeGenerateEditConfirm)
	}
	if model.input.targetPath != "team/api/key" {
		t.Fatalf("target path = %q, want %q", model.input.targetPath, "team/api/key")
	}
}

func TestMovingCursorToEntryLoadsMaskedPreview(t *testing.T) {
	t.Parallel()

	service := &fakeService{maskedValue: "hidden-value"}
	root := tree.Build([]string{"team/api/key"})
	root.Children[0].Expanded = true
	root.Children[0].Children[0].Expanded = true

	model := Model{service: service, root: root}
	model.refresh()

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("Update returned nil cmd")
	}

	msg := cmd()
	previewMsg, ok := msg.(previewLoadedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want previewLoadedMsg", msg)
	}
	if previewMsg.path != "team/api/key" {
		t.Fatalf("preview path = %q, want %q", previewMsg.path, "team/api/key")
	}
	if service.maskedPath != "team/api/key" {
		t.Fatalf("masked path = %q, want %q", service.maskedPath, "team/api/key")
	}
}

func TestEnterOnEntryStartsEditing(t *testing.T) {
	t.Parallel()

	service := &fakeService{}
	root := tree.Build([]string{"team/api/key"})
	root.Children[0].Expanded = true
	root.Children[0].Children[0].Expanded = true

	model := Model{service: service, root: root}
	model.refresh()
	model.focusPath("team/api/key")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("Update returned nil cmd")
	}
	if model.status != "editing team/api/key" {
		t.Fatalf("status = %q, want %q", model.status, "editing team/api/key")
	}
	if service.editPath != "team/api/key" {
		t.Fatalf("edit path = %q, want %q", service.editPath, "team/api/key")
	}
}

func TestSelectingEntryLoadsMaskedPreviewForNextCurrentEntry(t *testing.T) {
	t.Parallel()

	service := &fakeService{maskedValue: "******"}
	root := tree.Build([]string{"team/api/key", "team/api/next"})
	root.Children[0].Expanded = true
	root.Children[0].Children[0].Expanded = true

	model := Model{service: service, root: root, selected: map[string]bool{}, cut: map[string]bool{}}
	model.refresh()
	model.focusPath("team/api/key")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected preview load cmd after selecting entry")
	}
	if !model.selected["team/api/key"] {
		t.Fatalf("selected = %v, want current entry selected", model.selected)
	}
	if model.currentNode() == nil || model.currentNode().Path != "team/api/next" {
		t.Fatalf("current node = %v, want next entry", model.currentNode())
	}
	msg := cmd()
	previewMsg, ok := msg.(previewLoadedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want previewLoadedMsg", msg)
	}
	if previewMsg.path != "team/api/next" {
		t.Fatalf("preview path = %q, want %q", previewMsg.path, "team/api/next")
	}
}

func TestViewEmptyStoreShowsCreationHelp(t *testing.T) {
	t.Parallel()

	model := Model{root: tree.Build(nil), width: 40}

	view := model.View()

	if !strings.Contains(view, "Empty store. Create an entry to get started.") {
		t.Fatalf("view = %q, want empty-store message", view)
	}
	if !strings.Contains(view, "Search secrets") || !strings.Contains(view, "Store status") {
		t.Fatalf("view = %q, want panel layout labels", view)
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

	service := &fakeService{maskedValue: "******"}
	root := tree.Build([]string{"personal/website/toto/titi", "work/api/token"})

	model := Model{service: service, root: root}
	model.refresh()
	model.beginSearch()
	model.input.value = "titi"
	model.searchQuery = model.input.value
	model.applySearchFilter()

	cmd := model.handleSearchInput(tea.KeyMsg{Type: tea.KeyEnter})

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
	if cmd == nil {
		t.Fatal("expected preview load cmd after search selection")
	}
	msg := cmd()
	previewMsg, ok := msg.(previewLoadedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want previewLoadedMsg", msg)
	}
	if previewMsg.path != "personal/website/toto/titi" {
		t.Fatalf("preview path = %q, want selected path", previewMsg.path)
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

	if !strings.Contains(view, "Navigation") || !strings.Contains(view, "Entries") {
		t.Fatalf("view = %q, want detailed help", view)
	}
	if !strings.Contains(view, "n             New entry") || !strings.Contains(view, "r             Regenerate current password") {
		t.Fatalf("view = %q, want new and regenerate shortcuts", view)
	}
	if !strings.Contains(view, "R             Rename or move current entry") {
		t.Fatalf("view = %q, want rename shortcut", view)
	}
	if !strings.Contains(view, "? hide help • / search • enter edit • n new • R rename • q quit") {
		t.Fatalf("view = %q, want compact footer hint", view)
	}
}

func TestViewShowsModernPanelsAndPersistentSearchField(t *testing.T) {
	t.Parallel()

	root := tree.Build([]string{"team/api/key"})
	root.Children[0].Expanded = true
	root.Children[0].Children[0].Expanded = true

	model := Model{root: root, width: 100, height: 24}
	model.refresh()

	view := model.View()
	if !strings.Contains(view, "Search secrets") {
		t.Fatalf("view = %q, want persistent search panel", view)
	}
	if !strings.Contains(view, "Preview") || !strings.Contains(view, "Store status") {
		t.Fatalf("view = %q, want preview and status panels", view)
	}
	if !strings.Contains(view, "/ to search secrets") {
		t.Fatalf("view = %q, want search placeholder", view)
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
