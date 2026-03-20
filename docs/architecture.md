# Architecture

## Overview

`gp-tui` is a terminal UI written in Go. It does not implement password-store logic itself. Instead, it shells out to `gopass` and renders the result in a Bubble Tea interface.

The current application flow is synchronous:

1. Start the app.
2. Unlock the store through an interactive `gopass show -- <first-entry>` when the store is not empty.
3. Run `gopass sync` before entering the TUI.
4. Load the list of `gopass` entries.
5. Build a navigable tree from flat paths.
6. Render the visible portion of the tree.
7. React to keyboard input for navigation, preview, reveal, selection, search, delete, and copy.

## Package Layout

### `main`

`main.go` is the entrypoint. It creates the `gopass` service, initializes the UI model, and starts the Bubble Tea program in alt-screen mode.

### `internal/gopass`

`internal/gopass/service.go` contains the service interface used by the UI layer and the CLI-backed implementation.

Current responsibilities:

- list entries from `gopass`
- build interactive show commands for startup unlock
- build interactive sync commands for startup unlock
- show a secret
- show a masked preview
- build interactive edit commands for existing entries
- build interactive create commands for new entries
- copy a secret through `gopass`
- delete an entry through `gopass`

This package is the boundary between the TUI and the external `gopass` binary.

### `internal/tree`

`internal/tree/tree.go` converts flat entry paths into a directory tree.

Key responsibilities:

- build a hierarchical structure from paths such as `mail/personal/github`
- sort directories before leaf entries
- flatten only expanded branches for rendering

This keeps tree building and tree rendering concerns separate.

### `internal/ui`

The `internal/ui` package holds the Bubble Tea model and rendering code.

- `model.go` manages application state and key handling
- `view.go` renders the tree, preview area, and footer help
- `styles.go` defines Lip Gloss styles

The UI currently keeps the following state:

- loaded tree
- flattened visible nodes
- cursor position
- selected entries
- search query and temporary search state
- current preview text
- password visibility flag
- terminal size

The UI keeps creation and edit side effects in Bubble Tea commands, then reconciles the tree from `gopass` after the external editor exits.

## Execution Flow

### Startup

At startup, `main.go` first triggers an interactive unlock flow before creating the UI model:

- `gopass show -- <first-entry>` is used to unlock the store when at least one entry exists
- `gopass sync` runs immediately after, so SSH authentication prompts happen before the TUI starts

After that, `ui.NewModel` calls `service.List()`, then builds the tree with `tree.Build()`.

### Interaction Loop

Bubble Tea drives the interaction through the standard model lifecycle:

1. `Update()` receives key and window-size messages.
2. The model updates cursor position, expanded state, selection state, or preview state.
3. When needed, the model triggers Bubble Tea commands that call the `gopass` service.
4. `View()` renders the current tree slice, preview block, and help text.

### Create and Edit Behavior

- `e` launches `gopass edit` for the current entry
- `n` starts an inline prompt and then launches `gopass edit --create` for the submitted path
- after either editor flow completes, the model reloads the tree from `gopass`
- a successful create focuses the new entry and loads a masked preview
- the empty-store view still renders help text so the first entry can be created

### Delete Behavior

- `d` starts a lightweight confirmation prompt for the current entry or the selected entries
- confirming the prompt runs `gopass rm -f -- <path>` for each entry
- after the command finishes, the model reloads the tree from `gopass`
- partial failures keep the tree in sync and surface a status message

### Search Behavior

- `/` starts an inline search prompt
- search matches against the full entry path, case-insensitively
- while search is active, the tree is temporarily expanded so nested entries remain discoverable
- `enter` exits search and restores a normal tree view focused on the selected result
- `esc` cancels search and restores the previous expansion state

### Preview Behavior

When the current node is a file entry:

- `enter`, `l`, or `right` loads a masked preview
- `p` toggles between masked and full content
- `c` copies the current entry to the clipboard via `gopass`

Preview text is cleared when navigation changes the current context.

## Current Limitations

The current codebase is intentionally small and focused. Based on the code today, it does not yet include:

- asynchronous loading or refresh
- generate flows
- configuration management

## Tests

The project now includes focused unit tests for startup, creation, deletion, and search flows.

- `main_test.go` verifies the startup unlock and sync sequence
- `internal/gopass/service_test.go` verifies the `gopass` command wiring used by startup and editing flows
- `internal/ui/input_test.go` verifies inline creation, delete confirmation, search behavior, and the empty-store rendering path

These tests avoid the real password store by using a fake service and harmless subprocesses instead of mutating `gopass` data.
