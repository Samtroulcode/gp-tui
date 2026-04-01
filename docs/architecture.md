# Architecture

## Overview

`gp-tui` is a terminal UI written in Go. It does not implement password-store logic itself. Instead, it shells out to `gopass` and renders the result in a Bubble Tea interface.

The current application flow is synchronous:

1. Start the app.
2. Unlock the store through an interactive `gopass show -- <first-entry>` when the store is not empty.
3. Run `gopass sync` before entering the TUI.
4. Load the list of `gopass` entries.
5. Build a navigable tree from flat paths.
6. Render the three-panel interface.
7. React to keyboard input for navigation, preview, reveal, selection, search, delete, rename, generation, and copy.

## UI Layout

The application now renders three panels:

- **Explorer** on the left
- **Preview** on the right
- **Store status** at the bottom

The explorer panel is the primary interaction surface. It contains the tree and an always-visible `Search secrets` field. Even though the search field is always visible, it only becomes active when the user presses `/`.

The preview panel is informational. It shows either entry preview content or directory details. The status panel is also informational. It shows the current state summary, a short recent history of status messages, and placeholder areas for future store metadata such as `mounts`, `gpg`, and `git`.

There is no focus switching between panels. User navigation always stays in the explorer tree.

When help is opened with `?`, the UI renders a modal overlay on top of the existing application instead of replacing part of the layout.

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
- `view.go` composes the full layout and overlays the help modal
- `view_sections.go` renders the explorer, preview, status, and help sections
- `styles.go` defines Lip Gloss styles
- `search.go` manages local search state and filtering behavior

The UI currently keeps the following state:

- loaded tree
- flattened visible nodes
- cursor position
- selected entries
- cut entries
- search query and temporary search state
- current preview text
- password visibility flag
- status summary and recent status history
- help modal visibility
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
2. The model updates cursor position, expanded state, selection state, search state, help visibility, or preview state.
3. When needed, the model triggers Bubble Tea commands that call the `gopass` service.
4. `View()` renders the explorer, preview, and status panels, then optionally overlays the help modal.

### Create and Edit Behavior

- `e` launches `gopass edit` for the current entry
- `n` starts an inline prompt and then launches `gopass edit --create` for the submitted path
- after either editor flow completes, the model reloads the tree from `gopass`
- a successful create focuses the new entry and loads a masked preview
- the empty-store view still renders status and help hints so the first entry can be created

### Delete Behavior

- `d` starts a lightweight confirmation prompt for the current entry or the selected entries
- confirming the prompt runs `gopass rm -f -- <path>` for each entry
- after the command finishes, the model reloads the tree from `gopass`
- partial failures keep the tree in sync and surface a status message

### Search Behavior

- the `Search secrets` field is always visible in the explorer panel
- `/` activates that field and starts local search
- search matches against the full entry path, case-insensitively
- while search is active, the tree is temporarily expanded so nested entries remain discoverable
- `enter` exits search, clears the query, and restores a normal tree view focused on the selected result
- `esc` cancels search and restores the previous expansion state
- the current UI is intentionally structured to support future optional `fzf` integration without changing the panel layout

### Preview Behavior

When the current node is a file entry:

- moving the cursor onto the entry loads a masked preview automatically
- `enter` launches `gopass edit` for that entry
- `l` or `right` can be used to refresh the masked preview explicitly
- `p` toggles between masked and full content
- `c` copies the current entry to the clipboard via `gopass`

When the current node is a directory, the preview panel shows directory metadata instead of secret content.

Preview text is cleared when navigation changes the current context.

### Help Behavior

- `?` toggles a modal help overlay
- the modal does not move focus away from the explorer tree model
- when the modal is closed, the application returns to the same underlying layout and selection state

## Current Limitations

The current codebase is intentionally small and focused. Based on the code today, it does not yet include:

- asynchronous loading or refresh
- optional `fzf` search integration
- live backend store metadata for `mounts`, `gpg`, or `git`
- configuration management

## Tests

The project now includes focused unit tests for startup, creation, deletion, search, and help rendering flows.

- `main_test.go` verifies the startup unlock and sync sequence
- `internal/gopass/service_test.go` verifies the `gopass` command wiring used by startup and editing flows
- `internal/ui/input_test.go` verifies inline creation, delete confirmation, search behavior, help toggling, and the empty-store rendering path

These tests avoid the real password store by using a fake service and harmless subprocesses instead of mutating `gopass` data.
