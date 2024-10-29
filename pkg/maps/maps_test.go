package maps

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestDrop(t *testing.T) {
	g := NewGomegaWithT(t)

	origin := map[interface{}]interface{}{
		"key1": map[interface{}]interface{}{
			"key11": map[interface{}]interface{}{
				"key111": map[interface{}]interface{}{
					"key1111": "value1111",
				},
			},
			"key12": map[interface{}]interface{}{
				"key121": "value121",
			},
		},
		"key2": map[interface{}]interface{}{
			"key21": "value21",
		},
		"key3": "value3",
	}
	toDrop := map[interface{}]interface{}{
		"key1": map[interface{}]interface{}{
			"key11": map[interface{}]interface{}{
				"key111": map[interface{}]interface{}{
					"key1111": "value1111",
				},
			},
		},
		"key2": "value2",
	}

	expectedResult := map[interface{}]interface{}{
		"key1": map[interface{}]interface{}{
			"key12": map[interface{}]interface{}{
				"key121": "value121",
			},
		},
		"key3": "value3",
	}

	result := Drop(origin, toDrop)

	g.Expect(result).To(Equal(expectedResult))
}

func TestUnion(t *testing.T) {
	g := NewGomegaWithT(t)

	left := map[interface{}]interface{}{
		"key1": map[interface{}]interface{}{
			"key11": map[interface{}]interface{}{
				"key111": map[interface{}]interface{}{
					"key1111": "value1111",
				},
			},
			"key12": map[interface{}]interface{}{
				"key121": "value121",
			},
		},
		"key2": map[interface{}]interface{}{
			"key21": "value21",
		},
		"key3": "value3",
	}
	right := map[interface{}]interface{}{
		"key1": map[interface{}]interface{}{
			"key11": map[interface{}]interface{}{
				"key111": map[interface{}]interface{}{
					// overwrite value with value
					"key1111": 42,
				},
			},
			"key12": map[interface{}]interface{}{
				// overwrite value with map
				"key121": map[interface{}]interface{}{
					"key1212": "value1212",
				},
			},
			// add key/value
			"key13": "value13",
		},
	}
	expectedResult := map[interface{}]interface{}{
		"key1": map[interface{}]interface{}{
			"key11": map[interface{}]interface{}{
				"key111": map[interface{}]interface{}{
					"key1111": 42,
				},
			},
			"key12": map[interface{}]interface{}{
				"key121": map[interface{}]interface{}{
					"key1212": "value1212",
				},
			},
			"key13": "value13",
		},
		"key2": map[interface{}]interface{}{
			"key21": "value21",
		},
		"key3": "value3",
	}

	result := Union(left, right)

	g.Expect(result).To(Equal(expectedResult))
}
