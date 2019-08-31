package tree

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Tree []*File

type File struct {
	Path  string
	Depth int
	Files []*File
}

func New(root string, maxDepth int) (Tree, error) {
	return gather(root, maxDepth, 0)
}

func (tree Tree) JSON() (string, error) {
	jason, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jason), nil
}

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
