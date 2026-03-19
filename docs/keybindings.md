# Keybindings

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
| `enter` | Open the current directory or load a masked preview |
| `l` / `right` | Same as `enter` |
| `h` / `left` | Collapse the current directory or move back to the parent directory |
| `tab` | Expand or collapse all directories |

## Entry Actions

| Key | Action |
| --- | --- |
| `space` | Toggle selection on the current entry |
| `x` | Cut the selected entries, or the current entry when nothing is selected |
| `v` | Paste cut entries into the current directory |
| `n` | Start inline input to create a directory |
| `c` | Copy the current entry through `gopass show -c` |
| `p` | Toggle password visibility in the preview |

## Quit

| Key | Action |
| --- | --- |
| `q` | Quit the application |
| `ctrl+c` | Quit the application |

## Notes

- Selection applies only to file entries, not directories.
- Cut state applies only to file entries, not directories.
- `v` pastes into the current directory. When the cursor is on an entry, its parent directory is used.
- `n` opens an inline prompt. Press `enter` to create the directory or `esc` to cancel.
- Preview actions do nothing when the current node is a directory.
- After selecting an entry with `space`, the cursor advances to the next visible item when possible.
