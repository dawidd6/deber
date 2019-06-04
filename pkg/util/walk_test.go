package util_test

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWalk(t *testing.T) {
	err := util.Walk(
		"./",
		11,
		func(node util.Node) {
			fmt.Println(node.Depth, node.Path)
		},
	)

	assert.NoError(t, err)
}
