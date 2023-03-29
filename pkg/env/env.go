package env

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	PortHealthcheck = "port.healthcheck"
	PortMetrics     = "port.metrics"
	PortDebug       = "port.debug"

	SecretLabelSelector     = "secret.selector.label"
	SecretNameSelector      = "secret.selector.name"
	SecretNamespaceSelector = "secret.selector.namespace"
	SecretContentSelector   = "secret.selector.content"

	SecretFileSingle          = "secret.file.single"
	SecretFileNamePattern     = "secret.file.name.pattern"
	SecretFilePropertyPattern = "secret.file.property.pattern"

	SecretKeyTransformation = "secret.key.transformation"

	CallbackMethod      = "callback.method"
	CallbackUrl         = "callback.url"
	CallbackBody        = "callback.body"
	CallbackContenttype = "callback.contenttype"

	LogJson  = "log.json"
	LogLevel = "log.level"

	DefaultPortHealthcheck = 8383
	DefaultPortMetrics     = 8080
	DefaultPortDebug       = 1234

	DefaultLogJson  = false
	DefaultLogLevel = logrus.InfoLevel
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
