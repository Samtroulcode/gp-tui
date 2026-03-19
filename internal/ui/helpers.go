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
