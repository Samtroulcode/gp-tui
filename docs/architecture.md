# Architecture

## Overview

`gp-tui` is a terminal UI written in Go on top of Bubble Tea.

It is intentionally narrow in scope: the application does **not** implement password-store logic itself. Instead, it delegates store operations to the `gopass` CLI and focuses on rendering, navigation, and interaction.

That design keeps responsibilities clear:

- `gopass` remains the source of truth
- `gp-tui` handles UI state and user workflows
- store actions are executed through `gopass` commands, not reimplemented in Go

## High-level flow

The current application flow is:

1. Start the application.
2. Run the startup unlock flow through `gopass` when the store is not empty.
3. Run `gopass sync` before entering the TUI.
4. Load all entry paths from `gopass`.
5. Build a tree from flat paths.
6. Render the three-panel interface.
7. React to user input for navigation and store actions.

## UI layout

The current UI has three panels:

- **Explorer** on the left
- **Preview** on the right
- **Store status** at the bottom

### Explorer

The explorer is the only interactive panel.

It contains:

- the store tree
- the persistent **`Search secrets`** field

The search field is always visible, but it only becomes active when the user presses `/`.

### Preview

The preview panel is read-only.

Its behavior depends on the current node:

- when the current node is an **entry**, a **masked preview** is loaded automatically on selection
- when the current node is a **directory**, the panel shows directory information instead of secret content

The preview is informational only. Focus never moves into this panel.

### Store status

The status panel is also read-only.

It shows:

- the current status summary
- a short recent history of status messages
- reserved space for future store metadata

## Interaction model

There is no focus switching between panels.

The tree remains the main interaction surface for the whole application. The preview and status sections reflect state, but they are not directly editable.

The help screen opened with `?` is rendered as a **centered modal overlay** on top of the current layout.

## Current behavior

### Navigation

- `j` / `k` or arrow keys move the cursor
- `g` / `G` jump to the top or bottom
- `h` / `left` collapse a directory or move to its parent
- `l` / `right` expand a directory or refresh the current entry preview
- `tab` toggles all directories

### Entry behavior

- selecting an entry loads a masked preview automatically
- `enter` edits the current entry
- `c` copies the current entry through `gopass show -c`
- `p` toggles masked vs full preview
- `e` explicitly opens the current entry in `gopass edit`
- `r` starts password regeneration
- `R` starts rename or move
- `d` starts delete confirmation

### Directory behavior

- `enter` expands or collapses the current directory
- directories are not directly editable
- preview actions do not reveal secret content for directories

### Search behavior

Search is local to the already loaded tree.

- the `Search secrets` field is always visible
- `/` activates the field
- filtering matches full entry paths, case-insensitively
- while search is active, the tree is temporarily expanded
- `enter` keeps the current result and returns to normal tree navigation
- `esc` cancels search and restores the previous tree expansion state

## Package layout

### `main`

`main.go` is the entrypoint.

It:

- creates the `gopass` service
- runs the startup bootstrap flow
- creates the UI model
- starts the Bubble Tea program in alt-screen mode

### `internal/gopass`

This package is the boundary between the TUI and the external `gopass` binary.

Current responsibilities include:

- listing entries
- building startup unlock and sync commands
- showing full entry content
- building masked previews
- opening entries in the editor
- creating entries
- generating or regenerating passwords
- copying entries to the clipboard
- moving entries
- deleting entries

### `internal/tree`

This package converts flat `gopass` paths into a navigable directory tree.

Current responsibilities include:

- building a hierarchical structure from flat paths
- sorting directories before leaf entries
- flattening only expanded branches for rendering

### `internal/ui`

This package contains the Bubble Tea application model and rendering logic.

It manages:

- the visible tree state
- cursor position
- selected entries
- cut state
- search state
- preview content and visibility
- status messages
- prompts and confirmation flows
- help modal visibility
- terminal dimensions

UI side effects are performed through Bubble Tea commands, then reconciled with a tree reload from `gopass`.

## gopass boundary

`gp-tui` delegates store operations to `gopass` instead of implementing them itself.

Examples of wrapped commands include:

- `gopass ls --flat`
- `gopass show -- <path>`
- `gopass show -c -- <path>`
- `gopass edit -- <path>`
- `gopass edit --create -- <path>`
- `gopass generate ...`
- `gopass mv -- <source> <destination>`
- `gopass rm -f -- <path>`

This keeps the application small and testable while preserving `gopass` as the operational backend.

## Current limitations

The current codebase is intentionally small and does not yet include:

- asynchronous loading
- external search integration such as `fzf`
- live store metadata in the status panel
- configuration management or custom keybindings

## Tests

The project includes focused unit tests around the current flows, including:

- startup bootstrap behavior
- `gopass` command wiring
- creation and generation flows
- regeneration confirmation
- rename and move behavior
- delete confirmation
- local search behavior
- help modal rendering
- empty-store behavior

These tests use fakes and harmless subprocesses instead of mutating a real password store.
