package env

import (
	"log/slog"
)

const (
	PodName = "pod.name"

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

	SecretDeletionWatch = "secret.deletion.watch"

	CallbackMethod      = "callback.method"
	CallbackURL         = "callback.url"
	CallbackBody        = "callback.body"
	CallbackContentType = "callback.content-type"

	LogJson  = "log.json"
	LogLevel = "log.level"

	DefaultPortHealthcheck = 8383
	DefaultPortMetrics     = 8080
	DefaultPortDebug       = 1234

	DefaultLogJson  = false
	DefaultLogLevel = slog.LevelInfo
)
