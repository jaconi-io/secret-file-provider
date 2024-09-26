package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func New(secret *corev1.Secret) *slog.Logger {
	return slog.With(
		"namespace", secret.Namespace,
		"name", secret.Name,
	)
}

func InitLogging() {
	var level slog.Level
	err := level.UnmarshalText([]byte(viper.GetString(env.LogLevel)))
	if err != nil {
		panic(fmt.Errorf("invalid log level %q: %w", viper.GetString(env.LogLevel), err))
	}

	options := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if viper.GetBool(env.LogJson) {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	slog.SetDefault(slog.New(handler))
	log.SetLogger(logr.FromSlogHandler(handler))
}

func init() {
	cobra.OnInitialize(InitLogging)
}
