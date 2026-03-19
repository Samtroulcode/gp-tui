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

func createDirectoryCmd(service gopass.Service, directoryPath string, expanded map[string]bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := service.Mkdir(ctx, directoryPath); err != nil {
			return treeUpdatedMsg{err: err}
		}

		root, err := loadTree(service, ctx, expanded)
		if err != nil {
			return treeUpdatedMsg{err: err}
		}

		return treeUpdatedMsg{root: root, focusPath: directoryPath, status: fmt.Sprintf("created folder %s", directoryPath)}
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

func loadTree(service gopass.Service, ctx context.Context, expanded map[string]bool) (*tree.Node, error) {
	paths, err := service.List(ctx)
	if err != nil {
		return nil, err
	}

	root := tree.Build(paths)
	applyExpandedState(root, expanded)
	return root, nil
}
