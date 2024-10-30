package maps

import (
	gomaps "maps"
)

// Union will merge two given maps recursive, where the values of the 'right' one will overwrite the ones
// from the 'left'.
func Union(left, right map[interface{}]interface{}) map[interface{}]interface{} {
	out := gomaps.Clone(left)
	for k, v := range right {
		// If you use map[string]interface{}, ok is always false here, because [yaml.Unmarshal] returns
		// map[interface{}]interface{}.
		if v, ok := v.(map[interface{}]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[interface{}]interface{}); ok {
					out[k] = Union(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// Drop will remove known entries from a given map recursively.
func Drop(origin, toRemove map[interface{}]interface{}) map[interface{}]interface{} {
	out := gomaps.Clone(origin)
	for k, v := range toRemove {
		if _, ok := out[k]; !ok {
			// not existing, don't need to care
			continue
		}
		if v, ok := v.(map[interface{}]interface{}); ok {
			// found v itself a map
			if outChild, ok := out[k].(map[interface{}]interface{}); ok {
				// it's a map
				child := Drop(outChild, v)
				if child == nil {
					delete(out, k)
				} else {
					out[k] = child
				}
			} else {
				// simple value in original map: just drop it
				delete(out, k)
			}
		} else {
			// found leaf, drop entry
			delete(out, k)
		}
	}
	if len(out) < 1 {
		return nil
	}
	return out
}
