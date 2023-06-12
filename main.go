package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"regexp"
	"runtime"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/jaconi-io/secret-file-provider/pkg/countingfinalizer"
	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/setup"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "secret-file-provider",
		Short: "Secret File Provider",
		Long:  "Operator like sidecar to copy K8s secret content into a predefined filesystem location.",
		Run: func(cmd *cobra.Command, args []string) {

			logrus.Infof("Go Version: %s", runtime.Version())
			logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)

			// Get a config to talk to the apiserver
			cfg, err := config.GetConfig()
			if err != nil {
				logrus.WithError(err).Fatal("failed to get config for apiserver")
			}

			var mgr manager.Manager
			// connecting to the k8s api server fails if an e.g. istio sidecar has not yet finished starting up
			retry(30, func() error {
				ns := env.GetSingleNamespace()
				if ns != "" {
					// if only one NS is defined, we attach that to manager, so that
					// we are able to use K8s roles instead of clusterroles
					mgr, err = manager.New(cfg, manager.Options{
						MetricsBindAddress:     ":" + viper.GetString(env.PortMetrics),
						HealthProbeBindAddress: ":" + viper.GetString(env.PortHealthcheck),
						Namespace:              ns,
					})
				} else {
					mgr, err = manager.New(cfg, manager.Options{
						MetricsBindAddress:     ":" + viper.GetString(env.PortMetrics),
						HealthProbeBindAddress: ":" + viper.GetString(env.PortHealthcheck),
					})
				}
				return err
			})

			go func() {
				// handler is registered by blank import of net/http/pprof
				logrus.Println(http.ListenAndServe("localhost:"+viper.GetString(env.PortDebug), nil))
			}()

			// Add default liveness and readiness probes.
			_ = mgr.AddHealthzCheck("ping", healthz.Ping)
			_ = mgr.AddReadyzCheck("ping", healthz.Ping)

			// register controller implementations
			setup.RegisterControllers(mgr)

			logrus.Info("Starting the Service")

			// Start the Service
			if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
				logrus.WithError(err).Fatal("Manager exited non-zero")
			}
			logrus.Info("Retrieved SIGTERM")

			cleanup(mgr)
			logrus.Info("cleanup completed")
		},
	}
	env.Bootstrap(rootCmd)
	rootCmd.Execute()
}

func retry(maxAttempts int, action func() error) {
	err := action()
	if err != nil {
		if maxAttempts > 0 {
			// retry
			logrus.WithError(err).Infof("will retry %d more times", maxAttempts)
			time.Sleep(time.Second)
			retry(maxAttempts-1, action)
		} else {
			// give up
			logrus.WithError(err).Fatal("give up")
		}
	}
}

func cleanup(mgr manager.Manager) {
	ctx := context.Background()
	listOptions := &client.ListOptions{}

	if viper.GetString(env.SecretLabelSelector) != "" {
		labelSelector, err := labels.Parse(viper.GetString(env.SecretLabelSelector))
		if err != nil {
			logrus.WithError(err).Error("cleanup failed due to invalid secret label selector")
			return
		}

		listOptions.LabelSelector = labelSelector
	}

	ns := env.GetSingleNamespace()
	if ns != "" {
		listOptions.Namespace = ns
	}

	secrets := &corev1.SecretList{}
	if err := mgr.GetClient().List(context.Background(), secrets, listOptions); err != nil {
		logrus.Error("cleanup failed", err)
		return
	}

	// Filter for name pattern, if configured.
	var accept func(corev1.Secret) bool
	nameSelector := viper.GetString(env.SecretNameSelector)
	if nameSelector != "" {
		regex := regexp.MustCompilePOSIX(nameSelector)
		accept = func(secret corev1.Secret) bool {
			return regex.MatchString(secret.Name)
		}
	} else {
		accept = func(corev1.Secret) bool {
			return true
		}
	}

	for _, secret := range secrets.Items {
		if !accept(secret) {
			continue
		}

		// Decrement the finalizer, for each secret.
		patch := client.StrategicMergeFrom(secret.DeepCopy())
		countingfinalizer.Decrement(&secret, env.FinalizerPrefix)
		if err := mgr.GetClient().Patch(ctx, &secret, patch); err != nil {
			logrus.Error("cleanup failed for secret ", err)
			continue
		}
	}
}
