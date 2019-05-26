package omap

type OrderedMap struct {
	keys   []string
	values []interface{}
}

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

func (om *OrderedMap) Has(key string) bool {
	for i := range om.keys {
		if om.keys[i] == key {
			return true
		}
	}

	return false
}

func (om *OrderedMap) Keys() []string {
	return om.keys
}

func (om *OrderedMap) Values() []interface{} {
	return om.values
}

func (om *OrderedMap) Count() int {
	return len(om.keys)
}

func (om *OrderedMap) DeleteAll() {
	om.keys = nil
	om.values = nil
}

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
