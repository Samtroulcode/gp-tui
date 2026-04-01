package ui

import (
	"fmt"
	"strings"

	"gopass-tui/internal/tree"
)

func parentDirectory(entryPath string) string {
	lastSlash := strings.LastIndex(entryPath, "/")
	if lastSlash == -1 {
		return ""
	}

	return entryPath[:lastSlash]
}

func joinPath(directoryPath, name string) string {
	if directoryPath == "" {
		return name
	}

	return directoryPath + "/" + name
}

func displayDirectory(path string) string {
	if path == "" {
		return "store root"
	}

	return path
}

func renamePromptPrefix(entryPath string) string {
	parent := parentDirectory(entryPath)
	if parent == "" {
		return ""
	}

	return parent + "/"
}

func remapPathSet(paths map[string]bool, sourcePath, destinationPath string, includeChildren bool) {
	if len(paths) == 0 {
		return
	}

	if paths[sourcePath] {
		delete(paths, sourcePath)
		paths[destinationPath] = true
	}

	if !includeChildren {
		return
	}

	prefix := sourcePath + "/"
	for currentPath := range paths {
		if !strings.HasPrefix(currentPath, prefix) {
			continue
		}

		suffix := strings.TrimPrefix(currentPath, prefix)
		delete(paths, currentPath)
		paths[joinPath(destinationPath, suffix)] = true
	}
}

func entryCountLabel(count int) string {
	if count == 1 {
		return "1 entry"
	}

	return fmt.Sprintf("%d entries", count)
}

func applyExpandedState(node *tree.Node, expanded map[string]bool) {
	for _, child := range node.Children {
		if child.IsDir {
			child.Expanded = expanded[child.Path]
			applyExpandedState(child, expanded)
		}
	}
}

func collectExpandedState(node *tree.Node) map[string]bool {
	expanded := make(map[string]bool)
	collectExpandedStateRecursive(node, expanded)
	return expanded
}

func collectExpandedStateRecursive(node *tree.Node, expanded map[string]bool) {
	if node == nil {
		return
	}

	for _, child := range node.Children {
		if child.IsDir {
			expanded[child.Path] = child.Expanded
			collectExpandedStateRecursive(child, expanded)
		}
	}
}

func markAllDirectoriesExpanded(node *tree.Node, expanded map[string]bool) {
	if node == nil {
		return
	}

	for _, child := range node.Children {
		if child.IsDir {
			expanded[child.Path] = true
			markAllDirectoriesExpanded(child, expanded)
		}
	}
}

func setExpandedRecursive(node *tree.Node, expanded bool) {
	if node == nil {
		return
	}

	for _, child := range node.Children {
		if child.IsDir {
			child.Expanded = expanded
			setExpandedRecursive(child, expanded)
		}
	}
}

func expandPath(node *tree.Node, targetPath string) bool {
	if node == nil {
		return false
	}

	for _, child := range node.Children {
		if child.Path == targetPath {
			return true
		}
		if child.IsDir && expandPath(child, targetPath) {
			child.Expanded = true
			return true
		}
	}

	return false
}
