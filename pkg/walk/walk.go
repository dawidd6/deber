package walk

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Func = func(node Node) bool

type Node struct {
	Path  string
	Depth int
	Nodes []Node
}

func (node Node) Base() string {
	return filepath.Base(node.Path)
}

func (node Node) Dir() string {
	return filepath.Dir(node.Path)
}

func Walk(root string, maxDepth int, fn Func) error {
	nodes, err := walk(root, maxDepth, 0)

	loop(nodes, fn)

	return err
}

func loop(nodes []Node, fn Func) {
	for _, node := range nodes {
		exit := fn(node)
		if exit {
			return
		}
		loop(node.Nodes, fn)
	}
}

func walk(root string, maxDepth int, currentDepth int) ([]Node, error) {
	rootNodes := make([]Node, 0)
	newNodes := make([]Node, 0)

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
		newNodes, err = walk(path, maxDepth, currentDepth)
		if err != nil {
			return nil, err
		}

		// Construct new node
		newNode := Node{
			Path:  path,
			Depth: currentDepth,
			Nodes: newNodes,
		}

		// Append a node to the rest
		rootNodes = append(rootNodes, newNode)
	}

	// Finally return all nodes
	return rootNodes, nil
}
