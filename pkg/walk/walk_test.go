package walk_test

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWalk(t *testing.T) {
	err := walk.Walk(
		"./",
		11,
		func(node walk.Node) {
			fmt.Println(node.Depth, node.Path)
		},
	)

	assert.NoError(t, err)
}
