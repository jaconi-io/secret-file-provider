package countingfinalizer

import (
	"fmt"
	"strconv"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Decrement accepts an Object and decrements (or removes) the provided finalizer.
func Decrement(o client.Object, prefix string) {
	i := get(o, prefix)
	if i < 0 {
		return
	}

	if i <= 1 {
		controllerutil.RemoveFinalizer(o, fmt.Sprintf("%s%d", prefix, i))
		return
	}

	controllerutil.RemoveFinalizer(o, fmt.Sprintf("%s%d", prefix, i))
	controllerutil.AddFinalizer(o, fmt.Sprintf("%s%d", prefix, i-1))
}

// Increment accepts an Object and increments (or adds) the provided finalizer.
func Increment(o client.Object, prefix string) {
	i := get(o, prefix)
	if i < 0 {
		controllerutil.AddFinalizer(o, fmt.Sprintf("%s%d", prefix, 1))
		return
	}

	controllerutil.RemoveFinalizer(o, fmt.Sprintf("%s%d", prefix, i))
	controllerutil.AddFinalizer(o, fmt.Sprintf("%s%d", prefix, i+1))
}

// get checks an Object that the provided counting finalizer is present and returns its value.
func get(o client.Object, prefix string) int {
	finalizers := o.GetFinalizers()
	for _, finalizer := range finalizers {
		if strings.HasPrefix(finalizer, prefix) {
			value := strings.TrimPrefix(finalizer, prefix)
			i, err := strconv.Atoi(value)
			if err != nil {
				// Apparently not out finalizer.
				continue
			}

			return i
		}
	}

	return -1
}
