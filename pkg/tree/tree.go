// Package tree
package tree

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Tree type represents a tree of files and directories
type Tree []*File

// File type represents single file and it's children if it's a directory
type File struct {
	// Path is the full path of file
	Path string
	// Depth is the depth of file relative to root directory
	Depth int
	// Files are the children files if it's a directory
	Files []*File
}

// New function creates new Tree struct for given root directory and maximum depth
func New(root string, maxDepth int) (Tree, error) {
	return gather(root, maxDepth, 0)
}

// JSON function returns indented JSON object of tree
func (tree Tree) JSON() (string, error) {
	jason, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jason), nil
}

// Walk function executes given walker function for every file in tree
func (tree Tree) Walk(walker func(file *File) error) error {
	return walk(tree, walker)
}

func walk(tree Tree, walker func(file *File) error) error {
	for i := range tree {
		err := walker(tree[i])
		if err != nil {
			return err
		}

		err = walk(tree[i].Files, walker)
		if err != nil {
			return err
		}
	}

	return nil
}

func gather(root string, maxDepth int, currentDepth int) ([]*File, error) {
	rootFiles := make([]*File, 0)
	newFiles := make([]*File, 0)

	// Maximum depth reached, end recursion
	if currentDepth == maxDepth {
		return nil, nil
	}

	// Return if not directory
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	// Read directory contents
	entities, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	// One step further
	currentDepth++

	// Loop over directory contents
	for _, entity := range entities {
		// Construct full path of one entity
		path := filepath.Join(root, entity.Name())

		// Go deeper
		newFiles, err = gather(path, maxDepth, currentDepth)
		if err != nil {
			return nil, err
		}

		// Construct new file
		newFile := &File{
			Path:  path,
			Depth: currentDepth,
			Files: newFiles,
		}

		// Append a file to the rest
		rootFiles = append(rootFiles, newFile)
	}

	// Finally return all files
	return rootFiles, nil
}
