package env

import (
	"github.com/sirupsen/logrus"
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

	FinalizerPrefix = "jaconi.io/secret-file-provider-"
)
