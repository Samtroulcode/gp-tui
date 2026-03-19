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
	listPaths  []string
}

func (f fakeService) List(context.Context) ([]string, error) { return f.listPaths, nil }
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
func (fakeService) Copy(context.Context, string) error         { return errors.New("not implemented") }
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
