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
	synced    bool
	syncErr   error
	calls     []string
}

func (s *startupService) List(context.Context) ([]string, error) { return s.listPaths, s.listErr }

func (s *startupService) ShowCommand(ctx context.Context, path string) *exec.Cmd {
	s.showPath = path
	s.calls = append(s.calls, "show")
	if s.showErr != nil {
		return exec.CommandContext(ctx, "false")
	}
	return exec.CommandContext(ctx, "true")
}

func (s *startupService) SyncCommand(ctx context.Context) *exec.Cmd {
	s.synced = true
	s.calls = append(s.calls, "sync")
	if s.syncErr != nil {
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

func (*startupService) Delete(context.Context, string) error { return nil }

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
	if !service.synced {
		t.Fatal("sync was not called")
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
	if !service.synced {
		t.Fatal("sync was not called")
	}
	if len(service.calls) != 2 || service.calls[0] != "show" || service.calls[1] != "sync" {
		t.Fatalf("calls = %v, want [show sync]", service.calls)
	}
}

func TestUnlockStoreReturnsListError(t *testing.T) {
	t.Parallel()

	service := &startupService{listErr: errors.New("list failed")}

	err := unlockStore(context.Background(), service)
	if err == nil {
		t.Fatal("unlockStore returned nil error")
	}
	if service.synced {
		t.Fatal("sync should not run after list error")
	}
}

func TestUnlockStoreReturnsShowError(t *testing.T) {
	t.Parallel()

	service := &startupService{listPaths: []string{"team/api"}, showErr: errors.New("show failed")}

	err := unlockStore(context.Background(), service)
	if err == nil {
		t.Fatal("unlockStore returned nil error")
	}
	if service.synced {
		t.Fatal("sync should not run after show error")
	}
}

func TestUnlockStoreReturnsSyncError(t *testing.T) {
	t.Parallel()

	service := &startupService{listPaths: []string{"team/api"}, syncErr: errors.New("sync failed")}

	err := unlockStore(context.Background(), service)
	if err == nil {
		t.Fatal("unlockStore returned nil error")
	}
	if !service.synced {
		t.Fatal("sync was not called")
	}
}
