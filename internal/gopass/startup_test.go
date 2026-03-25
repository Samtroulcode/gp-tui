package gopass

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type startupServiceStub struct {
	listPaths []string
	listErr   error
	showPath  string
	showFails bool
	synced    bool
	syncFails bool
	calls     []string
}

func (s *startupServiceStub) List(context.Context) ([]string, error) { return s.listPaths, s.listErr }

func (s *startupServiceStub) ShowCommand(ctx context.Context, path string) *exec.Cmd {
	s.showPath = path
	s.calls = append(s.calls, "show")
	if s.showFails {
		return helperCommand(ctx, tHelperCommandFail)
	}

	return helperCommand(ctx, tHelperCommandPass)
}

func (s *startupServiceStub) SyncCommand(ctx context.Context) *exec.Cmd {
	s.synced = true
	s.calls = append(s.calls, "sync")
	if s.syncFails {
		return helperCommand(ctx, tHelperCommandFail)
	}

	return helperCommand(ctx, tHelperCommandPass)
}

const (
	tHelperCommandPass = "pass"
	tHelperCommandFail = "fail"
)

func helperCommand(ctx context.Context, outcome string) *exec.Cmd {
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=TestHelperProcess", "--", outcome)
	command.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return command
}

func TestHelperProcess(t *testing.T) {
	t.Helper()

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	separator := -1
	for index, arg := range args {
		if arg == "--" {
			separator = index
			break
		}
	}
	if separator == -1 || separator == len(args)-1 {
		os.Exit(2)
	}

	switch args[separator+1] {
	case tHelperCommandPass:
		os.Exit(0)
	case tHelperCommandFail:
		os.Exit(1)
	default:
		os.Exit(2)
	}
}

func TestBootstrapStoreSkipsUnlockForEmptyStore(t *testing.T) {
	t.Parallel()

	service := &startupServiceStub{}

	if err := BootstrapStore(context.Background(), service, StartupIO{Stdout: io.Discard, Stderr: io.Discard, UnlockOut: io.Discard}); err != nil {
		t.Fatalf("BootstrapStore returned error: %v", err)
	}
	if service.showPath != "" {
		t.Fatalf("showPath = %q, want empty", service.showPath)
	}
	if !service.synced {
		t.Fatal("sync was not called")
	}
}

func TestBootstrapStoreUsesFirstEntry(t *testing.T) {
	t.Parallel()

	service := &startupServiceStub{listPaths: []string{"team/api", "team/db"}}

	if err := BootstrapStore(context.Background(), service, StartupIO{Stdout: io.Discard, Stderr: io.Discard, UnlockOut: io.Discard}); err != nil {
		t.Fatalf("BootstrapStore returned error: %v", err)
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

func TestBootstrapStoreReturnsListError(t *testing.T) {
	t.Parallel()

	service := &startupServiceStub{listErr: errors.New("list failed")}

	err := BootstrapStore(context.Background(), service, StartupIO{Stdout: io.Discard, Stderr: io.Discard, UnlockOut: io.Discard})
	if err == nil {
		t.Fatal("BootstrapStore returned nil error")
	}
	if !strings.HasPrefix(err.Error(), "list store: ") {
		t.Fatalf("error = %q, want list store prefix", err.Error())
	}
	if service.synced {
		t.Fatal("sync should not run after list error")
	}
}

func TestBootstrapStoreReturnsShowError(t *testing.T) {
	t.Parallel()

	service := &startupServiceStub{listPaths: []string{"team/api"}, showFails: true}

	err := BootstrapStore(context.Background(), service, StartupIO{Stdout: io.Discard, Stderr: io.Discard, UnlockOut: io.Discard})
	if err == nil {
		t.Fatal("BootstrapStore returned nil error")
	}
	if !strings.HasPrefix(err.Error(), "unlock store with \"team/api\": unlock store: ") {
		t.Fatalf("error = %q, want wrapped unlock prefix", err.Error())
	}
	if service.synced {
		t.Fatal("sync should not run after show error")
	}
}

func TestBootstrapStoreReturnsSyncError(t *testing.T) {
	t.Parallel()

	service := &startupServiceStub{listPaths: []string{"team/api"}, syncFails: true}

	err := BootstrapStore(context.Background(), service, StartupIO{Stdout: io.Discard, Stderr: io.Discard, UnlockOut: io.Discard})
	if err == nil {
		t.Fatal("BootstrapStore returned nil error")
	}
	if !strings.HasPrefix(err.Error(), "initial sync: sync store: ") {
		t.Fatalf("error = %q, want wrapped sync prefix", err.Error())
	}
	if !service.synced {
		t.Fatal("sync was not called")
	}
}
