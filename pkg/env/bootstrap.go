package env

import (
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Bootstrap(rootCmd *cobra.Command) {
	rootCmd.Flags().String(PodName, "", "the pods name")
	rootCmd.Flags().Uint32(PortHealthcheck, DefaultPortHealthcheck, "port the health endpoints bind to")
	rootCmd.Flags().Uint32(PortMetrics, DefaultPortMetrics, "port the controller runtime metrics endpoint binds to")
	rootCmd.Flags().Uint32(PortDebug, DefaultPortDebug, "port the go debug information are present on")
	rootCmd.Flags().Bool(LogJson, DefaultLogJson, "output logs in JSON format")
	rootCmd.Flags().String(LogLevel, DefaultLogLevel.String(), "log level")
	rootCmd.Flags().String(SecretLabelSelector, "", "secret labels to consider")
	rootCmd.Flags().String(SecretNameSelector, "", "secret name pattern to consider")
	rootCmd.Flags().String(SecretContentSelector, "", "secret content path to copy")
	rootCmd.Flags().String(SecretKeyTransformation, "", "transformation function for all secret keys")
	rootCmd.Flags().Bool(SecretDeletionWatch, false, "set to 'true' if secret deletion should be watched and therefore their content needs to be dropped from FS")
	rootCmd.Flags().Bool(SecretFileSingle, false, "set to 'true' if each secret key should get it's own file")
	rootCmd.Flags().String(SecretFileNamePattern, "", "target filename pattern")
	rootCmd.Flags().String(SecretFilePropertyPattern, "", "base property path in target file")
	rootCmd.Flags().String(CallbackURL, "", "URL to call with GET request for successful file updates")
	rootCmd.Flags().String(CallbackMethod, http.MethodGet, "method for callback URL, sent on file updates")
	rootCmd.Flags().String(CallbackBody, "", "body sent with callback on file updates")
	rootCmd.Flags().String(CallbackContentType, "application/json", "Content-Type header of callback requests")

	rootCmd.MarkFlagRequired(PodName)
	rootCmd.MarkFlagRequired(SecretFileNamePattern)

	viper.BindPFlags(rootCmd.Flags())

	cobra.OnInitialize(unmarkRequired(rootCmd))
}

func initConfig() {
	// Allow flags containing dashes / dots to be set by environment variables which use underscores instead of dashes /
	// dots.
	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)
}

func init() {
	cobra.OnInitialize(initConfig, viper.AutomaticEnv)
}

// unmarkRequired works around an issue with cobra and viper, where required flags - set via environment variables - are
// not recognized. See https://github.com/spf13/viper/issues/397
func unmarkRequired(cmd *cobra.Command) func() {
	return func() {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if viper.IsSet(f.Name) {
				cmd.Flags().SetAnnotation(f.Name, cobra.BashCompOneRequiredFlag, []string{"false"})
			}
		})
	}
}
