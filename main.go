package main

import (
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/setup"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "secret-injector",
		Short: "Secret Injector",
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
			// connecting to the k8s api server fails if an existing istio sidecar has not yet finished starting up
			retry(30, func() error {
				mgr, err = manager.New(cfg, manager.Options{
					MetricsBindAddress:     ":" + viper.GetString(env.PortMetrics),
					HealthProbeBindAddress: ":" + viper.GetString(env.PortHealthcheck),
				})
				return err
			})

			go func() {
				// handler is registered by blank import of net/http/pprof
				logrus.Println(http.ListenAndServe("localhost:"+viper.GetString(env.PortDebug), nil))
			}()

			// Add default liveness and readiness probes.
			_ = mgr.AddHealthzCheck("ping", healthz.Ping)
			_ = mgr.AddReadyzCheck("ping", healthz.Ping)

			setup.RegisterControllers(mgr)

			logrus.Info("Starting the Cmd.")

			// Start the Cmd
			if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
				logrus.WithError(err).Fatal("Manager exited non-zero")
			}
			logrus.Info("Retrieved SIGTERM")
		},
	}
	setup.InitFlags(rootCmd)
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
