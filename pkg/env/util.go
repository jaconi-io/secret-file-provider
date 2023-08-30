package env

import (
	"strings"

	"github.com/spf13/viper"
)

// GetSingleNamespace will return the name of a single namespace, if set. This will return
// an empty string if either no namespace or multiple ones are actively selected.
func GetSingleNamespace() string {
	ns := viper.GetString(SecretNamespaceSelector)
	if ns == "" || strings.Contains(ns, ",") {
		// multiple namespaces selected, return empty string
		return ""
	}
	return ns
}

// GetFinalizer returns the finalizer name. The finalizer name depends on the pod name.
func GetFinalizer() string {
	prefix := "jaconi.io/secret-file-provider-"
	pod := viper.GetString(PodName)

	// Kubernetes limits finalizer names to 63 characters. We use the tail of the pod name, as it contains the hash and
	// is therefore less prone to collisions.
	if len(prefix)+len(pod) > 63 {
		maxPodLen := 63 - len(prefix)
		return prefix + pod[len(pod)-maxPodLen:]
	}

	return prefix + pod
}
