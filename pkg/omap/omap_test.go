package omap_test

import (
	"github.com/dawidd6/deber/pkg/omap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTest() (*omap.OrderedMap, []string, []interface{}) {
	return new(omap.OrderedMap),
		[]string{
			"key1",
			"key2",
			"key3",
			"key4",
			"key5",
		}, []interface{}{
			"val1",
			"val2",
			"val3",
			"val4",
			"val5",
		}
}

func TestHas(t *testing.T) {
	om, keys, values := newTest()

	for i := range keys {
		om.Append(keys[i], values[i])
	}

	assert.Equal(t, true, om.Has(keys[0]))
	assert.Equal(t, false, om.Has("nah"))
}

func TestAppend(t *testing.T) {
	om, keys, values := newTest()

	for i := range keys {
		om.Append(keys[i], values[i])
		om.Append(keys[i], values[i])
	}

	assert.Equal(t, len(keys), om.Count())
}

func TestPrepend(t *testing.T) {
	om, keys, values := newTest()

	for i := range keys {
		om.Append(keys[i], values[i])
	}

	om.Prepend("key0", "val0")
	om.Prepend("key0", "val0")

	assert.Equal(t, om.Keys()[0], "key0")
	assert.Equal(t, om.Values()[0], "val0")
	assert.Equal(t, om.Count(), len(keys)+1)
}

func TestDelete(t *testing.T) {
	om, keys, values := newTest()

	for i := range keys {
		om.Append(keys[i], values[i])
	}

	om.Delete("key1", "key2")

	assert.Equal(t, om.Keys()[0], keys[2])
	assert.Equal(t, om.Count(), len(keys)-2)
}

func TestDeleteAll(t *testing.T) {
	om, keys, values := newTest()

	for i := range keys {
		om.Append(keys[i], values[i])
	}

	om.DeleteAll()

	assert.Equal(t, om.Count(), 0)
}

func TestDeleteAllExcept(t *testing.T) {
	om, keys, values := newTest()

	for i := range keys {
		om.Append(keys[i], values[i])
	}

	om.DeleteAllExcept("key1", "key2")

	assert.Equal(t, 2, om.Count())
	assert.Equal(t, keys[:2], om.Keys())
}

func TestInsertBefore(t *testing.T) {
	om, keys, values := newTest()

	for i := range keys {
		om.Append(keys[i], values[i])
	}

	om.InsertBefore(keys[2], "key", "val")
	om.InsertBefore(keys[0], "key", "val")

	t.Log(om.Keys())

	assert.Equal(t, keys[:2], om.Keys()[:2])
	assert.Equal(t, "key", om.Keys()[2])
	assert.Equal(t, keys[2:], om.Keys()[3:])
	assert.Equal(t, len(keys), om.Count()-1)
}

func TestInsertAfter(t *testing.T) {
	om, keys, values := newTest()

	for i := range keys {
		om.Append(keys[i], values[i])
	}

	om.InsertAfter(keys[2], "key", "val")
	om.InsertAfter(keys[0], "key", "val")

	t.Log(om.Keys())

	assert.Equal(t, keys[:3], om.Keys()[:3])
	assert.Equal(t, "key", om.Keys()[3])
	assert.Equal(t, keys[3:], om.Keys()[4:])
	assert.Equal(t, len(keys), om.Count()-1)
}
