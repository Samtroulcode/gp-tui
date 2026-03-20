package ui

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"gopass-tui/internal/tree"
)

type fakeService struct {
	createPath string
	deleted    []string
	deleteErrs map[string]error
	listPaths  []string
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

	if cmd == nil {
		t.Fatal("submitInput returned nil cmd")
	}
	if model.status != "creating entry team/api/new-secret" {
		t.Fatalf("status = %q, want %q", model.status, "creating entry team/api/new-secret")
	}
	if service.createPath != "team/api/new-secret" {
		t.Fatalf("create path = %q, want %q", service.createPath, "team/api/new-secret")
	}
	if model.input.mode != inputModeNone {
		t.Fatalf("input mode = %v, want %v", model.input.mode, inputModeNone)
	}
	if model.input.value != "" {
		t.Fatalf("input value = %q, want empty", model.input.value)
	}
}

func TestViewEmptyStoreShowsCreationHelp(t *testing.T) {
	t.Parallel()

	model := Model{root: tree.Build(nil), width: 40}

	view := model.View()

	if !strings.Contains(view, "Empty store. Create an entry to get started.") {
		t.Fatalf("view = %q, want empty-store message", view)
	}
	if !strings.Contains(view, "n new entry") {
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
