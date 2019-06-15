package util_test

import (
	"github.com/dawidd6/deber/pkg/util"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCopyDir(t *testing.T) {
	source := "/tmp/source"
	target := "/tmp/target"
	dir := "/dir"

	err := os.MkdirAll(source+dir, os.ModePerm)
	assert.NoError(t, err)

	err = os.MkdirAll(target, os.ModePerm)
	assert.NoError(t, err)

	err = util.CopyDir(source+dir, target+dir, true)
	assert.NoError(t, err)
}
