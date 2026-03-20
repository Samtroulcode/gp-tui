# gp-tui

A minimal terminal UI for browsing `gopass` entries with Go and Bubble Tea.

Current capabilities include:

- tree navigation for stores and entries
- startup unlock flow backed by `gopass show -- <first-entry>` and `gopass sync`
- masked and revealed entry previews
- entry creation backed by `gopass edit --create`
- entry editing through `gopass edit`
- entry deletion with confirmation
- local entry search on full paths through `/`
- multi-selection for entries
- cut and paste moves backed by `gopass mv`
- clipboard copy through `gopass show -c`
- an empty-state view that still allows creating the first entry
- automated unit tests for gopass integration and core UI flows

See `docs/architecture.md`, `docs/gopass-integration.md`, `docs/keybindings.md`, and `docs/contributing.md` for the project documentation.

## Roadmap

- rename entries with a smarter move flow built on `gopass mv`
- generate new passwords from the TUI
