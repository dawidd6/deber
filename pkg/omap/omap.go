// Package 'omap' implements an ordered map and basic operations on it.
package omap

// OrderedMap struct represents a custom map made of two slices.
type OrderedMap struct {
	keys   []string
	values []interface{}
}

// Append adds key with its value at the end of the map.
//
// If key already exists in map, only its value is updated.
func (om *OrderedMap) Append(key string, value interface{}) {
	for i := range om.keys {
		if key == om.keys[i] {
			om.values[i] = value
			return
		}
	}

	om.keys = append(om.keys, key)
	om.values = append(om.values, value)
}

// Prepend adds key with its value at the beginning of the map.
//
// If key already exists in map, only its value is updated.
func (om *OrderedMap) Prepend(key string, value interface{}) {
	for i := range om.keys {
		if key == om.keys[i] {
			om.values[i] = value
			return
		}
	}

	om.keys = append([]string{key}, om.keys...)
	om.values = append([]interface{}{value}, om.values...)
}

// InsertBefore adds key with its value before specified key in map.
//
// If key already exists in map, only its value is updated.
func (om *OrderedMap) InsertBefore(before, key string, value interface{}) {
	for i := range om.keys {
		if key == om.keys[i] {
			om.values[i] = value
			return
		}
	}

	for i := range om.keys {
		if om.keys[i] == before {
			om.keys = append(om.keys[:i], append([]string{key}, om.keys[i:]...)...)
			om.values = append(om.values[:i], append([]interface{}{value}, om.values[i:]...)...)
			return
		}
	}
}

// InsertAfter adds key with its value after specified key in map.
//
// If key already exists in map, only its value is updated.
func (om *OrderedMap) InsertAfter(after, key string, value interface{}) {
	for i := range om.keys {
		if key == om.keys[i] {
			om.values[i] = value
			return
		}
	}

	for i := range om.keys {
		if om.keys[i] == after {
			om.keys = append(om.keys[:i+1], append([]string{key}, om.keys[i+1:]...)...)
			om.values = append(om.values[:i+1], append([]interface{}{value}, om.values[i+1:]...)...)
			return
		}
	}
}

// Delete removes specified keys with their values from map.
func (om *OrderedMap) Delete(keys ...string) {
	for _, key := range keys {
		for i := range om.keys {
			if om.keys[i] == key {
				om.keys = append(om.keys[:i], om.keys[i+1:]...)
				om.values = append(om.values[:i], om.values[i+1:]...)
				break
			}
		}
	}
}

// Has checks if specified key exists in map.
func (om *OrderedMap) Has(key string) bool {
	for i := range om.keys {
		if om.keys[i] == key {
			return true
		}
	}

	return false
}

// Keys returns all existing keys in map.
func (om *OrderedMap) Keys() []string {
	return om.keys
}

// Values returns all existing values in map.
func (om *OrderedMap) Values() []interface{} {
	return om.values
}

// Count returns the current length of map.
func (om *OrderedMap) Count() int {
	return len(om.keys)
}

// DeleteAll removes all elements from map.
func (om *OrderedMap) DeleteAll() {
	om.keys = nil
	om.values = nil
}

// DeleteAllExcept remove all elements from maps except those specified.
func (om *OrderedMap) DeleteAllExcept(keys ...string) {
	trash := make([]string, 0)

	for i := range om.keys {
		found := false

		for _, key := range keys {
			if key == om.keys[i] {
				found = true
			}
		}

		if !found {
			trash = append(trash, om.keys[i])
		}
	}

	om.Delete(trash...)
}
