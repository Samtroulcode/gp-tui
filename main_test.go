package main

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"gopass-tui/internal/gopass"
)

type startupService struct {
	listPaths []string
	listErr   error
	showPath  string
	showErr   error
}

func (s *startupService) List(context.Context) ([]string, error) { return s.listPaths, s.listErr }

func (s *startupService) ShowCommand(ctx context.Context, path string) *exec.Cmd {
	s.showPath = path
	if s.showErr != nil {
		return exec.CommandContext(ctx, "false")
	}
	return exec.CommandContext(ctx, "true")
}

func (*startupService) Show(context.Context, string) (string, error) { return "", nil }

func (*startupService) ShowMasked(context.Context, string) (string, error) { return "", nil }

func (*startupService) EditCommand(ctx context.Context, path string) *exec.Cmd {
	return exec.CommandContext(ctx, "true")
}

func (*startupService) CreateCommand(ctx context.Context, path string) *exec.Cmd {
	return exec.CommandContext(ctx, "true")
}

func (*startupService) Copy(context.Context, string) error { return nil }

func (*startupService) Move(context.Context, string, string) error { return nil }

var _ gopass.Service = (*startupService)(nil)

func TestUnlockStoreSkipsEmptyStore(t *testing.T) {
	t.Parallel()

	service := &startupService{}

	if err := unlockStore(context.Background(), service); err != nil {
		t.Fatalf("unlockStore returned error: %v", err)
	}
	if service.showPath != "" {
		t.Fatalf("showPath = %q, want empty", service.showPath)
	}
}

func TestUnlockStoreUsesFirstEntry(t *testing.T) {
	t.Parallel()

	service := &startupService{listPaths: []string{"team/api", "team/db"}}

	if err := unlockStore(context.Background(), service); err != nil {
		t.Fatalf("unlockStore returned error: %v", err)
	}
	if service.showPath != "team/api" {
		t.Fatalf("showPath = %q, want %q", service.showPath, "team/api")
	}
}

func TestUnlockStoreReturnsListError(t *testing.T) {
	t.Parallel()

	service := &startupService{listErr: errors.New("list failed")}

	err := unlockStore(context.Background(), service)
	if err == nil {
		t.Fatal("unlockStore returned nil error")
	}
}

func TestUnlockStoreReturnsShowError(t *testing.T) {
	t.Parallel()

	service := &startupService{listPaths: []string{"team/api"}, showErr: errors.New("show failed")}

	err := unlockStore(context.Background(), service)
	if err == nil {
		t.Fatal("unlockStore returned nil error")
	}
}
