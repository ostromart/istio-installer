package util

import (
	"strings"
	"path/filepath"
)

// Path is a path in slice form.
type Path []string

// FromString converts a string path of form a/b/c to a string slice representation.
func FromString(path string) []string {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return strings.Split(path, "/")
}

// String converts a string slice path representation of form ["a", "b", "c"] to a string representation like "a/b/c".
func (p Path) String() string {
	return strings.Join(p, "/")
}

// Tree is a tree.
type Tree map[string]interface{}


