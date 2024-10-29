package env

import (
	"log/slog"
)

const (
	PodName = "pod.name"

	PortHealthcheck = "port.healthcheck"
	PortMetrics     = "port.metrics"
	PortDebug       = "port.debug"

	// K8s label selector
	SecretLabelSelector = "secret.selector.label"
	// K8s secret name selector
	SecretNameSelector = "secret.selector.name"
	// K8s namespace selector
	SecretNamespaceSelector = "secret.selector.namespace"
	// read only a specific field of the whole secret data
	SecretContentSelector = "secret.selector.content"

	// true, if all secrets should be contained by a single file
	SecretFileSingle = "secret.file.single"
	// pattern for secret file names
	SecretFileNamePattern = "secret.file.name.pattern"
	// pattern for a secret property prefix
	SecretFilePropertyPattern = "secret.file.property.pattern"

	// transformation function for (K8s secret) keys
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
