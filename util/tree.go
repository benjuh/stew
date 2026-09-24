package util

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type TreeNode struct {
	Name      string     `json:"name"`
	Directory bool       `json:"directory"`
	Children  []TreeNode `json:"children,omitempty"`
}

// BuildTree returns a structured, sorted tree for a directory.
// maxDepth of zero means unlimited depth.
func BuildTree(path string, maxDepth int) (TreeNode, error) {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return TreeNode{}, err
	}
	if !info.IsDir() {
		return TreeNode{}, fmt.Errorf("path is not a directory: %s", path)
	}
	patterns, err := readTreeIgnore(filepath.Join(path, ".stewignore"))
	if err != nil {
		return TreeNode{}, err
	}
	return buildTreeNode(path, info.Name(), true, 0, maxDepth, patterns)
}

func buildTreeNode(path, name string, directory bool, depth, maxDepth int, patterns []string) (TreeNode, error) {
	node := TreeNode{Name: name, Directory: directory}
	if !directory || (maxDepth > 0 && depth >= maxDepth) {
		return node, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return TreeNode{}, err
	}
	for _, entry := range entries {
		if entry.Name() == ".stewignore" || ignoredTreeEntry(entry.Name(), entry.IsDir(), patterns) {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		child, err := buildTreeNode(filepath.Join(path, entry.Name()), entry.Name(), entry.IsDir(), depth+1, maxDepth, patterns)
		if err != nil {
			return TreeNode{}, err
		}
		node.Children = append(node.Children, child)
	}
	sort.Slice(node.Children, func(i, j int) bool {
		if node.Children[i].Directory != node.Children[j].Directory {
			return node.Children[i].Directory
		}
		return node.Children[i].Name < node.Children[j].Name
	})
	return node, nil
}

func PrintTreeNode(node TreeNode) {
	var builder strings.Builder
	builder.WriteString(formatTreeName(node))
	builder.WriteByte('\n')
	for index, child := range node.Children {
		writeTreeBranch(&builder, child, "", index == len(node.Children)-1)
	}
	fmt.Print(builder.String())
}

func writeTreeBranch(builder *strings.Builder, node TreeNode, prefix string, last bool) {
	branch := "├── "
	if last {
		branch = "└── "
	}
	builder.WriteString(prefix)
	builder.WriteString(branch)
	builder.WriteString(formatTreeName(node))
	builder.WriteByte('\n')
	childPrefix := prefix + "│   "
	if last {
		childPrefix = prefix + "    "
	}
	for index, child := range node.Children {
		writeTreeBranch(builder, child, childPrefix, index == len(node.Children)-1)
	}
}

func formatTreeName(node TreeNode) string {
	if node.Directory {
		return dirColor + node.Name + reset + "/"
	}
	return fileColor + node.Name + reset
}

func readTreeIgnore(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, filepath.ToSlash(line))
		}
	}
	return patterns, nil
}

func ignoredTreeEntry(name string, directory bool, patterns []string) bool {
	if directory {
		switch name {
		case ".git", ".hg", ".svn", "node_modules", "vendor":
			return true
		}
	}
	for _, pattern := range patterns {
		pattern = strings.TrimSuffix(pattern, "/")
		pattern = strings.TrimPrefix(pattern, "/")
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}
