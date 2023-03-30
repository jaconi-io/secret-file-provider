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
