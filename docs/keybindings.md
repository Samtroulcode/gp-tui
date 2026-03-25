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
| `/` | Start local search across full entry paths |

## Entry Actions

| Key | Action |
| --- | --- |
| `e` | Edit the current entry through `gopass edit` |
| `d` | Confirm deletion for the current entry or selected entries |
| `space` | Toggle selection on the current entry |
| `x` | Cut the selected entries, or the current entry when nothing is selected |
| `v` | Paste cut entries into the current directory |
| `n` | Start the inline new-entry flow |
| `r` | Regenerate the password for the current entry |
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
- `n` opens an inline prompt for the new entry path. After pressing `enter`, the TUI asks `Generate password? [y/N]`.
- Answer `n` or press `enter` at that first prompt to keep the editor-based flow with `gopass edit --create -- <path>`.
- Answer `y` to enter the unified generate wizard backed by `gopass generate`.
- The generate wizard collects the supported options in order: optional key, length, generator (`cryptic`, `memorable`, `xkcd`, `external`), symbols, strict, separator, language (`en`, `de`), clip, print, edit, commit message, and interactive commit.
- When `r` starts the same wizard for the current entry, it begins with a confirmation explaining that the current password will be overwritten.
- In regeneration mode, the wizard also asks whether to overwrite the entire secret with `--force-regen`.
- Press `esc` to cancel the current input step.
- `/` opens an inline search prompt. Type to filter on full entry paths, press `enter` to return to the normal tree centered on the selected result, or `esc` to cancel and restore the previous tree state.
- While search is active, the tree is temporarily expanded so nested entries can still be found.
- When possible, the `n` prompt is prefilled with the current directory path.
- The empty-store view still shows the footer help, so `n` can create the first entry.
- Preview actions do nothing when the current node is a directory.
- `e` and `c` do nothing when the current node is a directory.
- `r` does nothing when the current node is a directory.
- After selecting an entry with `space`, the cursor advances to the next visible item when possible.
- `d` opens a confirmation prompt. Press `y` or `enter` to delete, or `n` / `esc` to cancel.
