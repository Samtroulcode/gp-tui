package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ──────────────────────────────────────────────────────────────────

var (
	styleDir      = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true)
	styleEntry    = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	styleCursor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f5c2e7")).Bold(true)
	styleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true)
	styleHelp     = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70"))
	styleTitle    = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cba6f7")).
			Bold(true).
			Padding(0, 1)
	stylePreview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9399b2")).
			Padding(0, 2)
)

// ── Tree ────────────────────────────────────────────────────────────────────

type Node struct {
	Name     string
	Path     string // full gopass path
	Children []*Node
	IsDir    bool
	Expanded bool
}

func buildTree(paths []string) *Node {
	root := &Node{Name: "store", IsDir: true, Expanded: true}

	for _, p := range paths {
		parts := strings.Split(p, "/")
		cur := root
		for i, part := range parts {
			isLast := i == len(parts)-1
			found := false
			for _, c := range cur.Children {
				if c.Name == part {
					cur = c
					found = true
					break
				}
			}
			if !found {
				n := &Node{
					Name:  part,
					Path:  strings.Join(parts[:i+1], "/"),
					IsDir: !isLast,
				}
				cur.Children = append(cur.Children, n)
				cur = n
			}
		}
	}

	sortTree(root)
	return root
}

func sortTree(n *Node) {
	sort.Slice(n.Children, func(i, j int) bool {
		// dirs first, then alphabetical
		if n.Children[i].IsDir != n.Children[j].IsDir {
			return n.Children[i].IsDir
		}
		return n.Children[i].Name < n.Children[j].Name
	})
	for _, c := range n.Children {
		sortTree(c)
	}
}

// flatten visible nodes for rendering
type flatNode struct {
	node  *Node
	depth int
}

func flatten(n *Node, depth int) []flatNode {
	var out []flatNode
	for _, c := range n.Children {
		out = append(out, flatNode{node: c, depth: depth})
		if c.IsDir && c.Expanded {
			out = append(out, flatten(c, depth+1)...)
		}
	}
	return out
}

// ── Gopass commands ─────────────────────────────────────────────────────────

func gopassList() ([]string, error) {
	cmd := exec.Command("gopass", "ls", "--flat")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths, nil
}

func gopassShow(path string) string {
	cmd := exec.Command("gopass", "show", path)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("(error: %v)", err)
	}
	// mask first line (password)
	lines := strings.SplitN(string(out), "\n", 2)
	if len(lines) > 1 {
		return "●●●●●●●●\n" + lines[1]
	}
	return "●●●●●●●●"
}

// ── Model ───────────────────────────────────────────────────────────────────

type model struct {
	root     *Node
	visible  []flatNode
	cursor   int
	selected map[string]bool
	preview  string
	showPass bool
	width    int
	height   int
}

func initialModel() model {
	paths, err := gopassList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: gopass ls --flat failed: %v\n", err)
		os.Exit(1)
	}

	root := buildTree(paths)
	m := model{
		root:     root,
		selected: make(map[string]bool),
	}
	m.refresh()
	return m
}

func (m *model) refresh() {
	m.visible = flatten(m.root, 0)
	if m.cursor >= len(m.visible) {
		m.cursor = max(0, len(m.visible)-1)
	}
}

func (m *model) currentNode() *Node {
	if m.cursor >= 0 && m.cursor < len(m.visible) {
		return m.visible[m.cursor].node
	}
	return nil
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "j", "down":
			if m.cursor < len(m.visible)-1 {
				m.cursor++
				m.preview = ""
				m.showPass = false
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.preview = ""
				m.showPass = false
			}

		case "g":
			m.cursor = 0
			m.preview = ""

		case "G":
			m.cursor = max(0, len(m.visible)-1)
			m.preview = ""

		case "enter", "l", "right":
			if n := m.currentNode(); n != nil {
				if n.IsDir {
					n.Expanded = !n.Expanded
					m.refresh()
				} else {
					m.preview = gopassShow(n.Path)
					m.showPass = false
				}
			}

		case "h", "left":
			if n := m.currentNode(); n != nil {
				if n.IsDir && n.Expanded {
					n.Expanded = false
					m.refresh()
				} else {
					// jump to parent dir
					for i := m.cursor - 1; i >= 0; i-- {
						if m.visible[i].node.IsDir && m.visible[i].depth < m.visible[m.cursor].depth {
							m.cursor = i
							m.preview = ""
							break
						}
					}
				}
			}

		case "p":
			// toggle password visibility
			if m.preview != "" {
				m.showPass = !m.showPass
				if n := m.currentNode(); n != nil && !n.IsDir {
					if m.showPass {
						cmd := exec.Command("gopass", "show", n.Path)
						out, _ := cmd.Output()
						m.preview = string(out)
					} else {
						m.preview = gopassShow(n.Path)
					}
				}
			}

		case " ":
			if n := m.currentNode(); n != nil && !n.IsDir {
				m.selected[n.Path] = !m.selected[n.Path]
				if !m.selected[n.Path] {
					delete(m.selected, n.Path)
				}
				// move down after toggle
				if m.cursor < len(m.visible)-1 {
					m.cursor++
				}
			}

		case "c":
			// copy password to clipboard
			if n := m.currentNode(); n != nil && !n.IsDir {
				exec.Command("gopass", "show", "-c", n.Path).Run()
				m.preview = "✓ copied to clipboard"
			}

		case "tab":
			// expand/collapse all
			allExpanded := true
			for _, f := range m.visible {
				if f.node.IsDir && !f.node.Expanded {
					allExpanded = false
					break
				}
			}
			for _, f := range m.visible {
				if f.node.IsDir {
					f.node.Expanded = !allExpanded
				}
			}
			m.root.Expanded = true // root always expanded
			m.refresh()
		}
	}

	return m, nil
}

func (m model) View() string {
	if len(m.visible) == 0 {
		return "Empty store. Press q to quit."
	}

	var b strings.Builder

	// title
	title := styleTitle.Render("  gopass")
	sel := ""
	if len(m.selected) > 0 {
		sel = styleCursor.Render(fmt.Sprintf(" [%d selected]", len(m.selected)))
	}
	b.WriteString(title + sel + "\n")
	b.WriteString(strings.Repeat("─", min(m.width, 60)) + "\n")

	// tree - calculate visible area
	treeHeight := m.height - 5 // title + separator + help + margins
	if m.preview != "" {
		treeHeight -= min(strings.Count(m.preview, "\n")+2, 8)
	}
	treeHeight = max(treeHeight, 5)

	// scroll offset
	start := 0
	if m.cursor >= treeHeight {
		start = m.cursor - treeHeight + 1
	}
	end := min(start+treeHeight, len(m.visible))

	for i := start; i < end; i++ {
		f := m.visible[i]
		indent := strings.Repeat("  ", f.depth)

		icon := "  "
		if f.node.IsDir {
			if f.node.Expanded {
				icon = " "
			} else {
				icon = " "
			}
		}

		name := f.node.Name
		line := indent + icon + name

		switch {
		case i == m.cursor && m.selected[f.node.Path]:
			b.WriteString(styleCursor.Render("▌") + styleSelected.Render(line))
		case i == m.cursor:
			if f.node.IsDir {
				b.WriteString(styleCursor.Render("▌") + styleDir.Render(line))
			} else {
				b.WriteString(styleCursor.Render("▌" + line))
			}
		case m.selected[f.node.Path]:
			b.WriteString(" " + styleSelected.Render("● "+indent+name))
		case f.node.IsDir:
			b.WriteString(" " + styleDir.Render(line))
		default:
			b.WriteString(" " + styleEntry.Render(line))
		}
		b.WriteString("\n")
	}

	// preview pane
	if m.preview != "" {
		b.WriteString("\n")
		previewLines := strings.Split(m.preview, "\n")
		maxLines := min(len(previewLines), 6)
		for _, l := range previewLines[:maxLines] {
			b.WriteString(stylePreview.Render(l) + "\n")
		}
		if len(previewLines) > maxLines {
			b.WriteString(stylePreview.Render(fmt.Sprintf("  ... +%d lines", len(previewLines)-maxLines)) + "\n")
		}
	}

	// help bar
	help := "j/k nav • enter open • space select • c copy • p reveal • tab expand • q quit"
	b.WriteString("\n" + styleHelp.Render(help))

	return b.String()
}

// ── Main ────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
