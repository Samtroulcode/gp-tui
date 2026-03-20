package ui

import (
	"context"
	"fmt"
	"path"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
	"gopass-tui/internal/tree"
)

func loadPreviewCmd(service gopass.Service, requestID int, entryPath string, showPass bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			preview string
			err     error
		)

		if showPass {
			preview, err = service.Show(ctx, entryPath)
		} else {
			preview, err = service.ShowMasked(ctx, entryPath)
		}

		return previewLoadedMsg{
			requestID: requestID,
			path:      entryPath,
			preview:   preview,
			showPass:  showPass,
			err:       err,
		}
	}
}

func copyEntryCmd(service gopass.Service, entryPath string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		return copyCompletedMsg{path: entryPath, err: service.Copy(ctx, entryPath)}
	}
}

func editEntryCmd(service gopass.Service, entryPath string) tea.Cmd {
	ctx := context.Background()
	command := service.EditCommand(ctx, entryPath)

	return tea.ExecProcess(command, func(err error) tea.Msg {
		return editCompletedMsg{path: entryPath, err: err}
	})
}

func createEntryCmd(service gopass.Service, entryPath string) tea.Cmd {
	ctx := context.Background()
	command := service.CreateCommand(ctx, entryPath)

	return tea.ExecProcess(command, func(err error) tea.Msg {
		return createEntryCompletedMsg{path: entryPath, err: err}
	})
}

func deleteEntriesCmd(service gopass.Service, paths []string, focusPath string, expanded map[string]bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		deletedPaths := make([]string, 0, len(paths))
		var firstErr error

		for _, path := range paths {
			if err := service.Delete(ctx, path); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}

			deletedPaths = append(deletedPaths, path)
		}

		status := fmt.Sprintf("deleted %s", entryCountLabel(len(deletedPaths)))
		if firstErr != nil {
			status = fmt.Sprintf("deleted %s, %d failed: %v", entryCountLabel(len(deletedPaths)), len(paths)-len(deletedPaths), firstErr)
		}
		if len(deletedPaths) == 0 && firstErr != nil {
			status = fmt.Sprintf("delete failed: %v", firstErr)
		}

		return deleteCompletedMsg{
			focusPath:  focusPath,
			status:     status,
			expanded:   expanded,
			clearPaths: deletedPaths,
		}
	}
}

func pasteCutEntriesCmd(service gopass.Service, paths []string, targetDir string, expanded map[string]bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		failed := make(map[string]bool)
		var (
			moved    int
			firstErr error
		)

		for _, sourcePath := range paths {
			destinationPath := joinPath(targetDir, path.Base(sourcePath))
			if destinationPath == sourcePath {
				continue
			}

			if err := service.Move(ctx, sourcePath, destinationPath); err != nil {
				failed[sourcePath] = true
				if firstErr == nil {
					firstErr = err
				}
				continue
			}

			moved++
		}

		root, err := loadTree(service, ctx, expanded)
		if err != nil {
			return treeUpdatedMsg{err: err}
		}

		status := fmt.Sprintf("moved %s to %s", entryCountLabel(moved), displayDirectory(targetDir))
		if firstErr != nil {
			status = fmt.Sprintf("moved %s, %d failed: %v", entryCountLabel(moved), len(failed), firstErr)
		}

		return treeUpdatedMsg{
			root:       root,
			focusPath:  targetDir,
			status:     status,
			cut:        failed,
			replaceCut: true,
		}
	}
}

func reloadTreeCmd(service gopass.Service, focusPath, status string, expanded map[string]bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		root, err := loadTree(service, ctx, expanded)
		if err != nil {
			return treeUpdatedMsg{err: err}
		}

		return treeUpdatedMsg{root: root, focusPath: focusPath, status: status}
	}
}

func loadTree(service gopass.Service, ctx context.Context, expanded map[string]bool) (*tree.Node, error) {
	paths, err := service.List(ctx)
	if err != nil {
		return nil, err
	}

	root := tree.Build(paths)
	applyExpandedState(root, expanded)
	return root, nil
}
