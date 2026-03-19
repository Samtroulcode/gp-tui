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

### `gopass show -c <path>`

Used by `CLIService.Copy()`.

Purpose:

- delegate clipboard handling to `gopass`
- keep clipboard behavior outside the TUI layer

The command is executed for its side effect. On success, the UI shows a confirmation message.

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
- `gopass show "path" failed`
- `gopass show -c "path" failed`

The UI displays those errors in the preview area when an operation fails.

## Current Design Notes

- the service contract used by the UI is small: `List`, `Show`, `ShowMasked`, and `Copy`
- commands are executed synchronously
- there is no explicit `context.Context` handling yet in the current implementation
- there is no sanitization layer for stderr or command output yet
