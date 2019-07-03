package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Func = func(file *File) bool

type File struct {
	os.FileInfo
	dir   string
	depth int
	files []*File
}

func (file *File) Dir() string {
	return file.dir
}

func (file *File) Depth() int {
	return file.depth
}

func Walk(root string, maxDepth int, fn Func) error {
	files, err := walk(root, maxDepth, 0)

	loop(files, fn)

	return err
}

func loop(files []*File, fn Func) {
	for _, file := range files {
		exit := fn(file)
		if exit {
			return
		}
		loop(file.files, fn)
	}
}

func walk(root string, maxDepth int, currentDepth int) ([]*File, error) {
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
		newFiles, err = walk(path, maxDepth, currentDepth)
		if err != nil {
			return nil, err
		}

		// Get info about current file
		info, err = os.Stat(path)
		if info == nil {
			return nil, err
		}

		// Construct new file
		newFile := &File{
			FileInfo: info,
			depth:    currentDepth,
			dir:      root,
			files:    newFiles,
		}

		// Append a file to the rest
		rootFiles = append(rootFiles, newFile)
	}

	// Finally return all files
	return rootFiles, nil
}
