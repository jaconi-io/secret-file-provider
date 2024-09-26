package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"runtime"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/setup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "secret-file-provider",
		Short: "Secret File Provider",
		Long:  "Operator like sidecar to copy K8s secret content into a predefined filesystem location.",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("go", "version", runtime.Version(), "os", runtime.GOOS, "arch", runtime.GOARCH)

			// Get a config to talk to the apiserver
			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to get config for apiserver: %w", err)
			}

			var mgr manager.Manager
			// connecting to the k8s api server fails if an e.g. istio sidecar has not yet finished starting up
			retry(30, func() error {
				ns := env.GetSingleNamespace()
				if ns != "" {
					// if only one NS is defined, we attach that to manager, so that
					// we are able to use K8s roles instead of clusterroles
					mgr, err = manager.New(cfg, manager.Options{
						Metrics: server.Options{
							BindAddress: ":" + viper.GetString(env.PortMetrics),
						},
						HealthProbeBindAddress: ":" + viper.GetString(env.PortHealthcheck),
						Cache: cache.Options{
							DefaultNamespaces: map[string]cache.Config{
								ns: {},
							},
						},
					})
				} else {
					mgr, err = manager.New(cfg, manager.Options{
						Metrics: server.Options{
							BindAddress: ":" + viper.GetString(env.PortMetrics),
						},
						HealthProbeBindAddress: ":" + viper.GetString(env.PortHealthcheck),
					})
				}
				return err
			})

			go func() {
				// handler is registered by blank import of net/http/pprof
				slog.Info("", "error", http.ListenAndServe("localhost:"+viper.GetString(env.PortDebug), nil))
			}()

			// Add default liveness and readiness probes.
			_ = mgr.AddHealthzCheck("ping", healthz.Ping)
			_ = mgr.AddReadyzCheck("ping", healthz.Ping)

			// register controller implementations
			setup.RegisterControllers(mgr)

			slog.Info("starting the service")

			// Start the Service
			if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
				return fmt.Errorf("manager exited with non-zero exit code: %w", err)
			}

			slog.Info("retrieved SIGTERM")
			cleanup(mgr)
			slog.Info("cleanup completed")

			return nil
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
			slog.Info(fmt.Sprintf("will retry %d more times", maxAttempts), "error", err)
			time.Sleep(time.Second)
			retry(maxAttempts-1, action)
		} else {
			// give up
			slog.Error("give up", "error", err)
			os.Exit(1)
		}
	}
}

func cleanup(mgr manager.Manager) {
	ctx := context.Background()
	listOptions := &client.ListOptions{}

	if viper.GetString(env.SecretLabelSelector) != "" {
		labelSelector, err := labels.Parse(viper.GetString(env.SecretLabelSelector))
		if err != nil {
			slog.Error("cleanup failed due to invalid secret label selector", "error", err)
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
		slog.Error("cleanup failed", "error", err)
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

	// Remove the finalizer from each secret.
	for _, secret := range secrets.Items {
		if !accept(secret) {
			continue
		}

		if _, err := controllerutil.CreateOrPatch(ctx, mgr.GetClient(), &secret, func() error {
			controllerutil.RemoveFinalizer(&secret, env.GetFinalizer())
			return nil
		}); err != nil {
			slog.Error("cleanup failed for secret", "error", err)
			continue
		}
	}
}
