# Keybindings

## Interaction Model

The explorer tree is the only interactive panel. The preview and status panels are read-only. There is no keyboard focus switching between panels.

The `Search secrets` field is always visible in the explorer panel, but it only becomes active when you press `/`.

## Navigation

| Key | Action |
| --- | --- |
| `j` / `down` | Move cursor down |
| `k` / `up` | Move cursor up |
| `g` | Jump to the first visible item |
| `G` | Jump to the last visible item |

## Tree Actions

| Key | Action |
| --- | --- |
| `enter` | Expand the current directory or edit the current entry |
| `l` / `right` | Expand the current directory or refresh the masked preview |
| `h` / `left` | Collapse the current directory or move back to the parent directory |
| `tab` | Expand or collapse all directories |
| `/` | Focus the persistent `Search secrets` field and start local search |
| `?` | Toggle the help modal overlay |

## Entry Actions

| Key | Action |
| --- | --- |
| `e` | Edit the current entry through `gopass edit` |
| `d` | Confirm deletion for the current entry or selected entries |
| `space` | Toggle selection on the current entry |
| `x` | Cut the selected entries, or the current entry when nothing is selected |
| `v` | Paste cut entries into the current directory |
| `n` | Start the inline new-entry flow |
| `r` | Regenerate the current entry password |
| `R` | Rename or move the current entry |
| `c` | Copy the current entry through `gopass show -c` |
| `p` | Toggle password visibility in the preview |

## Prompts and Search

| Key | Action |
| --- | --- |
| `enter` | Confirm the current prompt or keep the current search selection |
| `esc` | Cancel the current prompt or cancel search |
| `y` / `n` | Answer confirmation prompts |

## Quit

| Key | Action |
| --- | --- |
| `q` | Quit the application |
| `ctrl+c` | Quit the application |

## Notes

- Selection applies only to file entries, not directories.
- Cut state applies only to file entries, not directories.
- `v` pastes into the current directory. When the cursor is on an entry, its parent directory is used.
- `n` opens an inline prompt for the new entry path. After pressing `enter`, the TUI asks `Generate password? [y/N]`.
- Answer `n` or press `enter` at that first prompt to create the entry in the editor with `gopass edit --create -- <path>`.
- Answer `y` to continue to `Quick generation with recommended defaults? [Y/n]`.
- Quick generation uses generator `cryptic`, length `24`, `symbols=true`, and `strict=true`.
- Answer `n` at the quick-generation prompt to start the full generate wizard.
- The full generate wizard collects, in order: optional key, generator (`cryptic`, `memorable`, `xkcd`, `external`), length, then generator-specific options:
  - `cryptic` → `symbols`, `strict`
  - `xkcd` → `separator`, `language`
- `r` starts the same flow for the current entry, but begins with an overwrite confirmation because the current password will be replaced.
- After generating a new entry, the TUI opens `gopass edit -- <path>` automatically.
- After regenerating an entry, the TUI reloads the tree and masked preview, then asks `Edit <path> now? [y/N]`.
- When the cursor is on a file entry, the preview panel loads its masked preview automatically.
- `/` activates the always-visible `Search secrets` field.
- Search filters on full entry paths, case-insensitively.
- While search is active, the tree is temporarily expanded so nested entries can still be found.
- Press `enter` during search to keep the current result and return to the normal tree view focused on that entry.
- Press `esc` during search to cancel and restore the previous tree expansion state.
- The search UI is local today, but the layout is already prepared for future optional `fzf` integration.
- `?` opens help as a modal overlay above the current layout.
- When possible, the `n` prompt is prefilled with the current directory path.
- The empty-store view still shows status and help hints, so `n` can create the first entry.
- Preview actions do nothing when the current node is a directory.
- `e`, `c`, and `r` do nothing when the current node is a directory.
- After selecting an entry with `space`, the cursor advances to the next visible item when possible.
- `d` opens a confirmation prompt. Press `y` or `enter` to delete, or `n` / `esc` to cancel.
