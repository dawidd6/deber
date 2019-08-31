package util_test

import (
	"github.com/dawidd6/deber/pkg/utils"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompareMountsEqual(t *testing.T) {
	a := []mount.Mount{
		{
			Source:   "src1",
			Target:   "target1",
			ReadOnly: true,
		}, {
			Source: "src2",
			Target: "target2",
		}, {
			Source: "src3",
			Target: "target3",
		}, {
			Source:   "src4",
			Target:   "target4",
			ReadOnly: true,
		},
	}

	b := []mount.Mount{
		{
			Source: "src2",
			Target: "target2",
		}, {
			Source:   "src1",
			Target:   "target1",
			ReadOnly: true,
		}, {
			Source:   "src4",
			Target:   "target4",
			ReadOnly: true,
		}, {
			Source: "src3",
			Target: "target3",
		},
	}

	equal := utils.CompareMounts(a, b)
	assert.True(t, equal)
}

func TestCompareMountsNotEqual(t *testing.T) {
	a := []mount.Mount{
		{
			Source:   "src1",
			Target:   "target1",
			ReadOnly: true,
		}, {
			Source: "src2",
			Target: "target2",
		}, {
			Source: "src3",
			Target: "target3",
		}, {
			Source:   "src4",
			Target:   "target4",
			ReadOnly: true,
		},
	}

	b := []mount.Mount{
		{
			Source:   "src2",
			Target:   "target2",
			ReadOnly: true,
		}, {
			Source:   "src1",
			Target:   "target1",
			ReadOnly: true,
		}, {
			Source:   "src4",
			Target:   "target4",
			ReadOnly: true,
		}, {
			Source:   "src3",
			Target:   "target3",
			ReadOnly: true,
		},
	}

	equal := utils.CompareMounts(a, b)
	assert.True(t, !equal)
}

func TestCompareMountsDifferentSizes(t *testing.T) {
	a := []mount.Mount{
		{
			Source:   "src1",
			Target:   "target1",
			ReadOnly: true,
		}, {
			Source: "src2",
			Target: "target2",
		}, {
			Source: "src3",
			Target: "target3",
		}, {
			Source:   "src4",
			Target:   "target4",
			ReadOnly: true,
		},
	}

	b := []mount.Mount{
		{
			Source:   "src2",
			Target:   "target2",
			ReadOnly: true,
		}, {
			Source:   "src1",
			Target:   "target1",
			ReadOnly: true,
		}, {
			Source:   "src4",
			Target:   "target4",
			ReadOnly: true,
		},
	}

	equal := utils.CompareMounts(a, b)
	assert.True(t, !equal)
}
