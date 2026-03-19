package tree

import (
	"sort"
	"strings"
)

// Node represents a gopass entry in a navigable tree.
// Directories can contain children, while leaf nodes represent entries.
type Node struct {
	Name     string
	Path     string
	Children []*Node
	IsDir    bool
	Expanded bool
}

// FlatNode is a visible node paired with its indentation depth.
// It is used by the UI layer to render only expanded branches.
type FlatNode struct {
	Node  *Node
	Depth int
}

// Build creates a tree from flat gopass paths.
// Paths are split on '/', then inserted into a root directory named "store".
func Build(paths []string) *Node {
	root := &Node{Name: "store", IsDir: true, Expanded: true}

	for _, path := range paths {
		parts := strings.Split(path, "/")
		current := root

		for index, part := range parts {
			isLast := index == len(parts)-1
			child := findChild(current, part)
			if child == nil {
				child = &Node{
					Name:  part,
					Path:  strings.Join(parts[:index+1], "/"),
					IsDir: !isLast,
				}
				current.Children = append(current.Children, child)
			}

			current = child
		}
	}

	sortChildren(root)
	return root
}

// Flatten returns the nodes that should be visible in the UI.
// Only expanded directories contribute their descendants.
func Flatten(root *Node, depth int) []FlatNode {
	var visible []FlatNode

	for _, child := range root.Children {
		visible = append(visible, FlatNode{Node: child, Depth: depth})
		if child.IsDir && child.Expanded {
			visible = append(visible, Flatten(child, depth+1)...)
		}
	}

	return visible
}

func findChild(node *Node, name string) *Node {
	for _, child := range node.Children {
		if child.Name == name {
			return child
		}
	}

	return nil
}

func sortChildren(node *Node) {
	sort.Slice(node.Children, func(i, j int) bool {
		if node.Children[i].IsDir != node.Children[j].IsDir {
			return node.Children[i].IsDir
		}

		return node.Children[i].Name < node.Children[j].Name
	})

	for _, child := range node.Children {
		sortChildren(child)
	}
}
