package env

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Bootstrap(rootCmd *cobra.Command) {
	rootCmd.Flags().Uint32(PortHealthcheck, DefaultPortHealthcheck, "port the health endpoints bind to")
	viper.BindPFlag(PortHealthcheck, rootCmd.Flags().Lookup(PortHealthcheck))

	rootCmd.Flags().Uint32(PortMetrics, DefaultPortMetrics, "port the controller runtime metrics endpoint binds to")
	viper.BindPFlag(PortMetrics, rootCmd.Flags().Lookup(PortMetrics))

	rootCmd.Flags().Uint32(PortDebug, DefaultPortDebug, "port the go debug information are present on")
	viper.BindPFlag(PortDebug, rootCmd.Flags().Lookup(PortDebug))

	rootCmd.Flags().Bool(LogJson, DefaultLogJson, "output logs in JSON format")
	viper.BindPFlag(LogJson, rootCmd.Flags().Lookup(LogJson))

	rootCmd.Flags().String(LogLevel, DefaultLogLevel.String(), fmt.Sprintf("log level"))
	viper.BindPFlag(LogLevel, rootCmd.Flags().Lookup(LogLevel))

	rootCmd.Flags().String(SecretLabelSelector, "", fmt.Sprintf("secret labels to consider"))
	viper.BindPFlag(SecretLabelSelector, rootCmd.Flags().Lookup(SecretLabelSelector))

	rootCmd.Flags().String(SecretNameSelector, "", fmt.Sprintf("secret name pattern to consider"))
	viper.BindPFlag(SecretNameSelector, rootCmd.Flags().Lookup(SecretNameSelector))

	rootCmd.Flags().String(SecretContentSelector, "", fmt.Sprintf("secret content path to copy"))
	viper.BindPFlag(SecretContentSelector, rootCmd.Flags().Lookup(SecretContentSelector))

	rootCmd.Flags().String(SecretKeyTransformation, "", fmt.Sprintf("transformation function for all secret keys"))
	viper.BindPFlag(SecretKeyTransformation, rootCmd.Flags().Lookup(SecretKeyTransformation))

	rootCmd.Flags().Bool(SecretFileSingle, false, fmt.Sprintf("set to 'true' if each secret key should get it's own file"))
	viper.BindPFlag(SecretFileSingle, rootCmd.Flags().Lookup(SecretFileSingle))

	rootCmd.Flags().String(SecretFileNamePattern, "", fmt.Sprintf("target filename pattern"))
	viper.BindPFlag(SecretFileNamePattern, rootCmd.Flags().Lookup(SecretFileNamePattern))

	rootCmd.Flags().String(SecretFilePropertyPattern, "", fmt.Sprintf("base property path in target file"))
	viper.BindPFlag(SecretFilePropertyPattern, rootCmd.Flags().Lookup(SecretFilePropertyPattern))

	rootCmd.Flags().String(CallbackUrl, "", fmt.Sprintf("url to call with GET request for successful file updates"))
	viper.BindPFlag(CallbackUrl, rootCmd.Flags().Lookup(CallbackUrl))

	rootCmd.Flags().String(CallbackMethod, "GET", fmt.Sprintf("method for callback URL, sent on file updates"))
	viper.BindPFlag(CallbackMethod, rootCmd.Flags().Lookup(CallbackMethod))

	rootCmd.Flags().String(CallbackBody, "", fmt.Sprintf("body sent with callback on file updates"))
	viper.BindPFlag(CallbackBody, rootCmd.Flags().Lookup(CallbackBody))

	rootCmd.Flags().String(CallbackContenttype, "application/json", fmt.Sprintf("content-type of callback request body"))
	viper.BindPFlag(CallbackContenttype, rootCmd.Flags().Lookup(CallbackContenttype))
}

func initConfig() {
	// Allow flags containing dashes / dots to be set by environment variables which use underscores instead of dashes /
	// dots.
	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.AutomaticEnv()
}

func init() {
	cobra.OnInitialize(initConfig, viper.AutomaticEnv)
}
