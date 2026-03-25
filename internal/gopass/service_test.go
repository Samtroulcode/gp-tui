package gopass

import (
	"context"
	"reflect"
	"testing"
)

func TestCLIServiceEditCommand(t *testing.T) {
	t.Parallel()

	cmd := CLIService{}.EditCommand(context.Background(), "team/api")
	want := []string{"gopass", "edit", "--", "team/api"}

	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("EditCommand args = %v, want %v", cmd.Args, want)
	}
}

func TestCLIServiceCreateCommand(t *testing.T) {
	t.Parallel()

	cmd := CLIService{}.CreateCommand(context.Background(), "team/api")
	want := []string{"gopass", "edit", "--create", "--", "team/api"}

	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("CreateCommand args = %v, want %v", cmd.Args, want)
	}
}

func TestCLIServiceGenerateArgs(t *testing.T) {
	t.Parallel()

	args := CLIService{}.generateArgs("team/api", 24)
	want := []string{"generate", "--", "team/api", "24"}

	if !reflect.DeepEqual(args, want) {
		t.Fatalf("generateArgs = %v, want %v", args, want)
	}
}

func TestCLIServiceGenerateRejectsNonPositiveLength(t *testing.T) {
	t.Parallel()

	err := CLIService{}.Generate(context.Background(), "team/api", 0)
	if err == nil {
		t.Fatal("Generate returned nil error")
	}
	if err.Error() != "password length must be positive" {
		t.Fatalf("error = %q, want %q", err.Error(), "password length must be positive")
	}
}

func TestCLIServiceShowCommand(t *testing.T) {
	t.Parallel()

	cmd := CLIService{}.ShowCommand(context.Background(), "team/api")
	want := []string{"gopass", "show", "--", "team/api"}

	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("ShowCommand args = %v, want %v", cmd.Args, want)
	}
}

func TestCLIServiceSyncCommand(t *testing.T) {
	t.Parallel()

	cmd := CLIService{}.SyncCommand(context.Background())
	want := []string{"gopass", "sync"}

	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("SyncCommand args = %v, want %v", cmd.Args, want)
	}
}
