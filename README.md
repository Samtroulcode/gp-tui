# gp-tui

A lightweight terminal UI for `gopass`.

`gp-tui` is a small Go application built with Bubble Tea that wraps the `gopass` CLI in a focused terminal interface. It does not reimplement password-store logic, encryption, or clipboard handling. `gopass` remains the source of truth; `gp-tui` provides navigation, preview, and common store actions in a cleaner UI.

## Key ideas

- **Wrap `gopass`, don't reimplement it.**
- **Keep the UI small and predictable.**
- **Delegate store operations to `gopass` commands.**
- **Reflect the current store state instead of maintaining a separate model.**

## Features

- Three-panel layout:
  - **Explorer** on the left
  - **Preview** on the right
  - **Store status** at the bottom
- Tree-based browsing of stores, folders, and entries
- Persistent **`Search secrets`** field in the explorer
- Local search on full entry paths, activated with `/`
- Centered help modal with `?`
- Automatic **masked preview** when selecting an entry
- `enter` edits the current entry, or expands/collapses the current directory
- Entry creation flow
- Password generation and regeneration flows
- Rename and move support
- Multi-selection for entries
- Cut and paste backed by `gopass mv`
- Clipboard copy backed by `gopass show -c`
- Delete with confirmation
- Empty-store state that still allows creating the first entry

## Requirements

- Go **1.22+**
- [`gopass`](https://www.gopass.pw/) installed
- A working `gopass` setup on your machine

## Install or run

### Run from source

```bash
go run .
```

### Build locally

```bash
go build -o gp-tui .
./gp-tui
```

Before starting the TUI, make sure `gopass` works in your terminal.

## Usage

### Interaction model

`gp-tui` renders three panels, but only the **explorer tree** is interactive.

- The **preview panel** is read-only
- The **status panel** is read-only
- The **search field** is always visible, but only becomes active when you press `/`
- The **help screen** opens as a centered modal overlay

### Common actions

- `j` / `k` or arrow keys: move in the tree
- `enter`: expand/collapse a directory, or edit the current entry
- `/`: activate local search
- `c`: copy the current entry with `gopass show -c`
- `n`: create a new entry
- `r`: regenerate the current entry
- `R`: rename or move the current entry
- `d`: delete the current entry or current selection
- `?`: open help
- `q`: quit

See [`docs/keybindings.md`](docs/keybindings.md) for the full key reference.

## How it works with gopass

`gp-tui` does not manipulate the password store directly. It shells out to `gopass` for store operations such as:

- listing entries
- showing entry content
- copying to the clipboard
- editing and creating entries
- generating passwords
- moving entries
- deleting entries

See [`docs/gopass-integration.md`](docs/gopass-integration.md) for details.

## Architecture

The codebase is intentionally small:

- `main.go` boots the app and runs the startup unlock/sync flow
- `internal/gopass/` wraps `gopass` commands
- `internal/tree/` builds and flattens the entry tree
- `internal/ui/` contains the Bubble Tea model, rendering, prompts, and actions

See [`docs/architecture.md`](docs/architecture.md) for the full overview.

## Docs

- [`docs/architecture.md`](docs/architecture.md)
- [`docs/gopass-integration.md`](docs/gopass-integration.md)
- [`docs/keybindings.md`](docs/keybindings.md)
- [`docs/contributing.md`](docs/contributing.md)

## Roadmap

The current app already covers the core store workflow: browse, search, preview, create, edit, generate, regenerate, rename, move, copy, select, and delete.

Likely next steps:

- richer store metadata in the status panel
- optional search improvements
- configuration and customization
- packaging and distribution docs
- guided store setup and administration flows

## License

GPL-3.0
