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

	args, err := CLIService{}.generateArgs(GenerateRequest{Path: "team/api", Length: 24})
	if err != nil {
		t.Fatalf("generateArgs returned error: %v", err)
	}
	want := []string{"generate", "--generator", "cryptic", "--lang", "en", "--", "team/api", "24"}

	if !reflect.DeepEqual(args, want) {
		t.Fatalf("generateArgs = %v, want %v", args, want)
	}
}

func TestCLIServiceGenerateArgsIncludeAllOptions(t *testing.T) {
	t.Parallel()

	args, err := CLIService{}.generateArgs(GenerateRequest{
		Path:      "team/api",
		Key:       "password",
		Length:    32,
		Force:     true,
		Symbols:   true,
		Generator: "xkcd",
		Strict:    true,
		Separator: "-",
		Language:  "de",
	})
	if err != nil {
		t.Fatalf("generateArgs returned error: %v", err)
	}
	want := []string{
		"generate",
		"--force",
		"--symbols",
		"--generator", "xkcd",
		"--strict",
		"--sep", "-",
		"--lang", "de",
		"--", "team/api", "password", "32",
	}

	if !reflect.DeepEqual(args, want) {
		t.Fatalf("generateArgs = %v, want %v", args, want)
	}
}

func TestCLIServiceGenerateCommandRejectsInvalidRequest(t *testing.T) {
	t.Parallel()

	_, err := CLIService{}.GenerateCommand(context.Background(), GenerateRequest{Path: "team/api", Length: 0})
	if err == nil {
		t.Fatal("GenerateCommand returned nil error")
	}
	if err.Error() != "password length must be positive" {
		t.Fatalf("error = %q, want %q", err.Error(), "password length must be positive")
	}
}

func TestCLIServiceGenerateCommandBuildsCommand(t *testing.T) {
	t.Parallel()

	cmd, err := CLIService{}.GenerateCommand(context.Background(), GenerateRequest{Path: "team/api", Length: 24})
	if err != nil {
		t.Fatalf("GenerateCommand returned error: %v", err)
	}
	want := []string{"gopass", "generate", "--generator", "cryptic", "--lang", "en", "--", "team/api", "24"}

	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("GenerateCommand args = %v, want %v", cmd.Args, want)
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
