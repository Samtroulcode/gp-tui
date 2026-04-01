# Keybindings

## Interaction model

The explorer tree is the only interactive panel.

- The **preview** panel is read-only
- The **store status** panel is read-only
- The persistent **`Search secrets`** field only becomes active when you press `/`
- `?` opens a centered help modal overlay

## Navigation

| Key | Action |
| --- | --- |
| `j` / `down` | Move cursor down |
| `k` / `up` | Move cursor up |
| `g` | Jump to the first visible item |
| `G` | Jump to the last visible item |

## Tree actions

| Key | Action |
| --- | --- |
| `enter` | Expand/collapse the current directory, or edit the current entry |
| `l` / `right` | Expand the current directory, or refresh the masked preview for the current entry |
| `h` / `left` | Collapse the current directory, or move to its parent |
| `tab` | Expand or collapse all directories |
| `/` | Activate the persistent `Search secrets` field |
| `?` | Toggle the help modal |

## Entry actions

| Key | Action |
| --- | --- |
| `e` | Edit the current entry through `gopass edit` |
| `c` | Copy the current entry through `gopass show -c` |
| `p` | Toggle masked vs full preview |
| `n` | Start the new-entry flow |
| `r` | Regenerate the current entry password |
| `R` | Rename or move the current node |
| `d` | Delete the current entry or current selection |
| `space` | Toggle selection on the current entry |
| `x` | Cut selected entries, or the current entry if nothing is selected |
| `v` | Paste cut entries into the current directory |

## Prompts and search

| Key | Action |
| --- | --- |
| `enter` | Confirm the current prompt or keep the current search result |
| `esc` | Cancel the current prompt or cancel search |
| `y` / `n` | Answer confirmation prompts |

## Quit

| Key | Action |
| --- | --- |
| `q` | Quit the application |
| `ctrl+c` | Quit the application |

## Notes

- Selection applies only to entries, not directories.
- Cut state applies only to entries, not directories.
- When possible, the new-entry prompt is prefilled with the current directory path.
- The empty-store view still allows creating the first entry.
- When the cursor lands on an entry, its masked preview loads automatically.
- Preview actions do nothing when the current node is a directory.
- `e`, `c`, and `r` do nothing when the current node is a directory.
- `v` pastes into the current directory. If the cursor is on an entry, its parent directory is used.

## New entry flow

Press `n` to start entry creation.

1. Enter the target path.
2. The TUI asks: `Generate password? [y/N]`.

If you answer **no** or press `enter`, the app runs:

```bash
gopass edit --create -- <path>
```

If you answer **yes**, the app asks:

```text
Quick generation with recommended defaults? [Y/n]
```

Quick generation uses:

- generator: `cryptic`
- length: `24`
- symbols: enabled
- strict: enabled

If quick generation is declined, the full wizard collects:

1. optional key
2. generator: `cryptic`, `memorable`, `xkcd`, or `external`
3. length
4. generator-specific options

Generator-specific options:

- `cryptic` → `symbols`, `strict`
- `xkcd` → `separator`, `language`

After generating a new entry, the app opens it in `gopass edit`.

## Regeneration flow

Press `r` to regenerate the current entry.

- the flow starts with overwrite confirmation
- regeneration runs through `gopass generate`
- after success, the tree reloads and a masked preview is restored
- the app then asks whether to open the entry in `gopass edit`

## Search behavior

- `/` activates the visible `Search secrets` field
- search is local to the currently loaded tree
- matches are done against full entry paths, case-insensitively
- while search is active, the tree is temporarily expanded
- `enter` keeps the current result and returns to normal navigation
- `esc` cancels search and restores the previous expansion state

## Help modal

- `?` opens help as a centered modal overlay
- while help is open, `q`, `?`, or `esc` close it
- `ctrl+c` still quits the application
