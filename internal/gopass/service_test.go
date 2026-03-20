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

func TestCLIServiceShowCommand(t *testing.T) {
	t.Parallel()

	cmd := CLIService{}.ShowCommand(context.Background(), "team/api")
	want := []string{"gopass", "show", "--", "team/api"}

	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("ShowCommand args = %v, want %v", cmd.Args, want)
	}
}
