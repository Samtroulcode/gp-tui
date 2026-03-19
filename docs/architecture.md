# Architecture

## Overview

`gp-tui` is a terminal UI written in Go. It does not implement password-store logic itself. Instead, it shells out to `gopass` and renders the result in a Bubble Tea interface.

The current application flow is synchronous:

1. Start the app.
2. Load the list of `gopass` entries.
3. Build a navigable tree from flat paths.
4. Render the visible portion of the tree.
5. React to keyboard input for navigation, preview, reveal, selection, and copy.

## Package Layout

### `main`

`main.go` is the entrypoint. It creates the `gopass` service, initializes the UI model, and starts the Bubble Tea program in alt-screen mode.

### `internal/gopass`

`internal/gopass/service.go` contains the service interface used by the UI layer and the CLI-backed implementation.

Current responsibilities:

- list entries from `gopass`
- show a secret
- show a masked preview
- copy a secret through `gopass`

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
- current preview text
- password visibility flag
- terminal size

## Execution Flow

### Startup

At startup, `ui.NewModel` calls `service.List()`, then builds the tree with `tree.Build()`.

### Interaction Loop

Bubble Tea drives the interaction through the standard model lifecycle:

1. `Update()` receives key and window-size messages.
2. The model updates cursor position, expanded state, selection state, or preview state.
3. When needed, the model calls the `gopass` service directly.
4. `View()` renders the current tree slice, preview block, and help text.

### Preview Behavior

When the current node is a file entry:

- `enter`, `l`, or `right` loads a masked preview
- `p` toggles between masked and full content
- `c` copies the current entry to the clipboard via `gopass`

Preview text is cleared when navigation changes the current context.

## Current Limitations

The current codebase is intentionally small and focused. Based on the code today, it does not yet include:

- asynchronous loading or refresh
- search
- create, edit, delete, or generate flows
- configuration management
- automated tests
