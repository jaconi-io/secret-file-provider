package logger

import (
	"os"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

func New(secret *corev1.Secret) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"namespace": secret.Namespace,
		"name":      secret.Name,
	})
}

func InitLogging() {
	if viper.GetBool(env.LogJson) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	level, err := logrus.ParseLevel(viper.GetString(env.LogLevel))
	if err == nil {
		logrus.SetLevel(level)
	} else {
		logrus.SetLevel(env.DefaultLogLevel)
	}

	logrus.SetOutput(os.Stdout)
}

func init() {
	cobra.OnInitialize(InitLogging)
}
