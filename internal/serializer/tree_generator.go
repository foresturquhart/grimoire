package serializer

import (
	"path/filepath"
	"sort"
	"strings"
)

// TreeGenerator defines an interface for generating a directory tree representation.
type TreeGenerator interface {
	// GenerateTree takes a list of file paths and returns a tree structure representation.
	GenerateTree(filePaths []string) *TreeNode
}

// TreeNode represents a node in the directory tree.
type TreeNode struct {
	Name     string
	IsDir    bool
	Children []*TreeNode
}

// DefaultTreeGenerator is a concrete implementation of TreeGenerator that creates
// a tree structure from a list of file paths.
type DefaultTreeGenerator struct{}

// NewDefaultTreeGenerator returns a new instance of DefaultTreeGenerator.
func NewDefaultTreeGenerator() *DefaultTreeGenerator {
	return &DefaultTreeGenerator{}
}

// GenerateTree takes a slice of file paths and builds a tree structure.
// It returns the root node of the tree.
func (g *DefaultTreeGenerator) GenerateTree(filePaths []string) *TreeNode {
	if len(filePaths) == 0 {
		return &TreeNode{
			Name:  "",
			IsDir: true,
		}
	}

	// Sort the paths for consistent output
	sort.Strings(filePaths)

	// Build a tree structure
	root := &TreeNode{
		Name:     "",
		IsDir:    true,
		Children: []*TreeNode{},
	}

	// Map to track nodes by path for efficient lookups
	nodeMap := make(map[string]*TreeNode)
	nodeMap[""] = root

	// Add all files to the tree
	for _, path := range filePaths {
		parts := strings.Split(filepath.ToSlash(path), "/")
		currentPath := ""

		// Build the tree structure
		for i, part := range parts {
			isFile := i == len(parts)-1

			// Construct parent path and current path
			parentPath := currentPath
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			// Check if node already exists
			if _, exists := nodeMap[currentPath]; !exists {
				// Create new node
				newNode := &TreeNode{
					Name:     part,
					IsDir:    !isFile,
					Children: []*TreeNode{},
				}

				// Add to parent's children
				parentNode := nodeMap[parentPath]
				parentNode.Children = append(parentNode.Children, newNode)

				// Sort children by name
				sort.Slice(parentNode.Children, func(i, j int) bool {
					// Directories come before files
					if parentNode.Children[i].IsDir != parentNode.Children[j].IsDir {
						return parentNode.Children[i].IsDir
					}
					// Alphabetical order within same type
					return parentNode.Children[i].Name < parentNode.Children[j].Name
				})

				// Add to map
				nodeMap[currentPath] = newNode
			}
		}
	}

	return root
}
