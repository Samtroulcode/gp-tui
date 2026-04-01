# gp-tui

A minimal terminal UI for browsing `gopass` entries with Go and Bubble Tea.

Current capabilities include:

- a three-panel layout with an explorer on the left, a preview panel on the right, and a status panel at the bottom
- tree navigation for stores and entries, with user focus always kept in the explorer tree
- a persistent `Search secrets` field that is always visible in the explorer panel
- local entry search on full paths through `/`, with the UI already prepared for future optional `fzf` integration
- a help modal overlay toggled with `?`
- startup unlock flow backed by `gopass show -- <first-entry>` and `gopass sync`
- masked and revealed entry previews
- entry creation with a simplified flow: `n` asks for a path, then `Generate password? [y/N]`; declining opens `gopass edit --create -- <path>`, accepting starts password generation
- quick generation with recommended defaults: `cryptic`, length `24`, `symbols=true`, `strict=true`
- a full generate wizard when quick generation is declined: optional key, generator (`cryptic`, `memorable`, `xkcd`, `external`), length, and generator-specific options
- password regeneration for the current entry through `r`, with overwrite confirmation, tree/preview reload, and an optional post-generation edit prompt
- entry editing through `gopass edit`
- entry deletion with confirmation
- multi-selection for entries
- cut and paste moves backed by `gopass mv`
- clipboard copy through `gopass show -c`
- a status panel with a current summary, recent status history, and reserved slots for future store metadata (`mounts`, `gpg`, `git`)
- an empty-state view that still allows creating the first entry
- automated unit tests for gopass integration and core UI flows

See `docs/architecture.md`, `docs/gopass-integration.md`, `docs/keybindings.md`, and `docs/contributing.md` for the project documentation.

## Roadmap

The current product already covers tree navigation, preview, creation, generation, editing, deletion, multi-selection, move, local search, and the new three-panel UI. The next priorities are:

- **Core store management**
  - better search, including optional `fzf` integration
  - richer store metadata in the status panel
  - dedicated store administration flows for mounts and setup

- **Configuration and customization**
  - themes
  - custom store paths
  - configurable keybindings

- **Documentation and distribution**
  - a man page
  - TLDR page support / mention

- **Store administration and bootstrap**
  - initialize a new store from the app
  - manage mounts
  - guided `gopass` setup, including GPG and age configuration
