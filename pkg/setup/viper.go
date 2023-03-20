package setup

import (
	"fmt"
	"strings"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitFlags(rootCmd *cobra.Command) {
	rootCmd.Flags().Uint32(env.PortHealthcheck, env.DefaultPortHealthcheck, "port the health endpoints bind to")
	viper.BindPFlag(env.PortHealthcheck, rootCmd.Flags().Lookup(env.PortHealthcheck))

	rootCmd.Flags().Uint32(env.PortMetrics, env.DefaultPortMetrics, "port the controller runtime metrics endpoint binds to")
	viper.BindPFlag(env.PortMetrics, rootCmd.Flags().Lookup(env.PortMetrics))

	rootCmd.Flags().Uint32(env.PortDebug, env.DefaultPortDebug, "port the go debug information are present on")
	viper.BindPFlag(env.PortDebug, rootCmd.Flags().Lookup(env.PortDebug))

	rootCmd.Flags().Bool(env.LogJson, env.DefaultLogJson, "output logs in JSON format")
	viper.BindPFlag(env.LogJson, rootCmd.Flags().Lookup(env.LogJson))

	rootCmd.Flags().String(env.LogLevel, env.DefaultLogLevel.String(), fmt.Sprintf("log level"))
	viper.BindPFlag(env.LogLevel, rootCmd.Flags().Lookup(env.LogLevel))

	rootCmd.Flags().String(env.SecretLabelSelector, "", fmt.Sprintf("secret labels to consider"))
	viper.BindPFlag(env.SecretLabelSelector, rootCmd.Flags().Lookup(env.SecretLabelSelector))

	rootCmd.Flags().String(env.SecretNameSelector, "", fmt.Sprintf("secret name pattern to consider"))
	viper.BindPFlag(env.SecretNameSelector, rootCmd.Flags().Lookup(env.SecretNameSelector))

	rootCmd.Flags().String(env.SecretContentSelector, "", fmt.Sprintf("secret content path to copy"))
	viper.BindPFlag(env.SecretContentSelector, rootCmd.Flags().Lookup(env.SecretContentSelector))

	rootCmd.Flags().String(env.SecretKeyTransformation, "", fmt.Sprintf("transformation function for all secret keys"))
	viper.BindPFlag(env.SecretKeyTransformation, rootCmd.Flags().Lookup(env.SecretKeyTransformation))

	rootCmd.Flags().Bool(env.SecretFileSingle, false, fmt.Sprintf("set to 'true' if each secret key should get it's own file"))
	viper.BindPFlag(env.SecretFileSingle, rootCmd.Flags().Lookup(env.SecretFileSingle))

	rootCmd.Flags().String(env.SecretFileNamePattern, "", fmt.Sprintf("target filename pattern"))
	viper.BindPFlag(env.SecretFileNamePattern, rootCmd.Flags().Lookup(env.SecretFileNamePattern))

	rootCmd.Flags().String(env.SecretFilePropertyPattern, "", fmt.Sprintf("base property path in target file"))
	viper.BindPFlag(env.SecretFilePropertyPattern, rootCmd.Flags().Lookup(env.SecretFilePropertyPattern))

	rootCmd.Flags().String(env.CallbackUrl, "", fmt.Sprintf("url to call with GET request for successful file updates"))
	viper.BindPFlag(env.CallbackUrl, rootCmd.Flags().Lookup(env.CallbackUrl))

	rootCmd.Flags().String(env.CallbackMethod, "GET", fmt.Sprintf("method for callback URL, sent on file updates"))
	viper.BindPFlag(env.CallbackMethod, rootCmd.Flags().Lookup(env.CallbackMethod))

	rootCmd.Flags().String(env.CallbackBody, "", fmt.Sprintf("body sent with callback on file updates"))
	viper.BindPFlag(env.CallbackBody, rootCmd.Flags().Lookup(env.CallbackBody))

	rootCmd.Flags().String(env.CallbackContenttype, "application/json", fmt.Sprintf("content-type of callback request body"))
	viper.BindPFlag(env.CallbackContenttype, rootCmd.Flags().Lookup(env.CallbackContenttype))
}

func initConfig() {
	// Allow flags containing dashes / dots to be set by environment variables which use underscores instead of dashes /
	// dots.
	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.AutomaticEnv()
}

func init() {
	cobra.OnInitialize(initConfig, logger.InitLogging, viper.AutomaticEnv)
}
