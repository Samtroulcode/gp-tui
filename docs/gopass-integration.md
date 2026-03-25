# gopass Integration

## Overview

`gp-tui` integrates with `gopass` by invoking the CLI through `os/exec` in `internal/gopass/service.go`.

The application treats `gopass` as the source of truth. It does not parse or manipulate the password store directly.

## Current Commands

### `gopass ls --flat`

Used by `CLIService.List()`.

Purpose:

- fetch all entries as flat paths
- provide the initial dataset used to build the in-memory tree

The output is trimmed, split by line, and converted to a `[]string`.

### `gopass show <path>`

Used by `CLIService.Show()`.

Purpose:

- load the full content of an entry
- support password reveal in the preview area

The raw command output is returned as a string.

### `gopass show -- <path>`

Used by `CLIService.ShowCommand()`.

Purpose:

- trigger the startup unlock flow before Bubble Tea starts
- let `gopass` surface any interactive unlock prompt directly in the terminal

The command is run as an interactive process and its standard output is discarded so the secret is not printed during startup.

### `gopass show -c <path>`

Used by `CLIService.Copy()`.

Purpose:

- delegate clipboard handling to `gopass`
- keep clipboard behavior outside the TUI layer

The command is executed for its side effect. On success, the UI shows a confirmation message.

### `gopass sync`

Used by `CLIService.SyncCommand()`.

Purpose:

- run a sync immediately after startup unlock
- surface SSH authentication prompts before the TUI opens

The command is run as an interactive process during startup.

### `gopass edit <path>`

Used by `CLIService.EditCommand()`.

Purpose:

- open an existing entry in the user's editor
- keep editing delegated to `gopass`

The UI runs the command as an interactive process. On success, it reloads the tree and refreshes the preview for the edited entry.

### `gopass edit --create <path>`

Used by `CLIService.CreateCommand()`.

Purpose:

- create a new entry when the user declines password generation in the `n` flow
- keep entry creation delegated to `gopass`
- allow the first entry to be created even when the store view is empty

The `n` flow always starts by collecting the entry path. The UI then asks `Generate password? [y/N]`. When the user answers `n` or presses `enter`, it runs `gopass edit --create -- <path>`.

The UI runs the command as an interactive process. On success, it reloads the tree, focuses the new entry, and loads a masked preview.

### `gopass generate ... -- <path> [key] <length>`

Used by `CLIService.GenerateCommand()` through a structured `GenerateRequest`.

Purpose:

- create a new entry from the `n` flow when the user answers `y` to `Generate password? [y/N]`
- regenerate the password for the current entry from `r`
- keep password generation delegated to `gopass`
- avoid reimplementing generation logic in the TUI by treating `gopass generate` as the source of truth

Both entry creation and password regeneration use the same wizard. The UI collects the request fields and passes them to `GenerateCommand()`, which validates and builds the final `gopass generate` command line.

Wizard flow:

1. For `n`, ask `Generate password? [y/N]`.
2. For `r`, ask for overwrite confirmation because the current password will be replaced.
3. Collect the supported `gopass generate` options exposed by the TUI:
   - optional key
   - length
   - generator: `cryptic`, `memorable`, `xkcd`, `external`
   - symbols
   - strict
   - force-regen
   - separator
   - language: `en`, `de`
   - clip
   - print
   - edit
   - commit message
   - interactive commit

Defaults currently set by the UI are length `24`, generator `cryptic`, and language `en`.

For regeneration, the request always includes `--force` so `gopass` may overwrite the existing entry. The wizard also exposes `--force-regen` to replace the entire secret instead of only the password line.

On success, the UI reloads the tree, focuses the new entry, and loads a masked preview.

### `gopass mv <source> <destination>`

Used by `CLIService.Move()`.

Purpose:

- move an entry to another directory in the store
- back the TUI cut and paste workflow without reimplementing store operations

The command is executed for its side effect. On success, the UI reloads the tree from `gopass`.

### `gopass rm -f -- <path>`

Used by `CLIService.Delete()`.

Purpose:

- remove an entry from the store
- back the TUI delete confirmation flow without reimplementing store operations

The command is executed for its side effect. On success, the UI reloads the tree from `gopass`.

## Masked Preview Strategy

`CLIService.ShowMasked()` is implemented on top of `Show()`.

Behavior:

1. Load the full entry content.
2. Split it into at most two parts.
3. Replace the first line with `********`.
4. Keep the remaining lines unchanged.

This lets the UI show metadata stored below the password line while hiding the secret by default.

## Error Handling

Errors from command execution are wrapped with command context, for example:

- `gopass ls --flat failed`
- `gopass show path failed`
- `gopass show -c path failed`
- `gopass edit path failed`
- `gopass edit --create path failed`
- `gopass generate ... -- path [key] length failed`
- `gopass mv source destination failed`
- `gopass rm -f -- path failed`

The UI displays those errors in the preview area when an operation fails.

## Automated Test Coverage

The current test suite focuses on creation safety and command wiring without touching a real password store.

- `main_test.go` covers the startup `show` then `sync` sequence
- `internal/gopass/service_test.go` checks the command arguments built for startup, edit, create, and structured generate operations
- `internal/ui/input_test.go` covers inline entry creation input, the unified generate wizard, regeneration confirmation, delete confirmation, local search behavior, empty-state rendering, and validation errors

These tests use fakes and harmless subprocesses, so they do not modify the user's existing `gopass` store.

## Current Design Notes

- the service contract used by the UI is small: `List`, `ShowCommand`, `SyncCommand`, `Show`, `ShowMasked`, `EditCommand`, `CreateCommand`, `GenerateCommand`, `Copy`, `Delete`, and `Move`
- command execution accepts `context.Context` and uses `exec.CommandContext`
- stdout and stderr are handled separately so warnings do not pollute successful command output
- Bubble Tea side effects are triggered through commands and messages, then the tree is reloaded from `gopass`
- search is local to the already loaded store tree; it does not call `gopass find`
