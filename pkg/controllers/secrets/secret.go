package secrets

import (
	"strings"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/templates"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

// readSecretContent reads the content of the given secret as key value pairs into a map. Note that
// this will also apply the secret content selector, so not every key of the original secret might
// be represented in the resulting map.
func readSecretContent(secret *corev1.Secret) map[interface{}]interface{} {
	selectorTemplate := viper.GetString(env.SecretContentSelector)
	mapContent := make(map[interface{}]interface{})
	content := ""
	if len(selectorTemplate) < 1 {
		// return all
		for k, v := range secret.Data {
			mapContent[k] = string(v)
		}
	} else if !strings.Contains(selectorTemplate, "{{") {
		// not a go template, log warning and return all
		logrus.Warnf("Illegal selector pattern '%s'. Expecting go template", selectorTemplate)
		for k, v := range secret.Data {
			mapContent[k] = string(v)
		}
	} else {
		content = templates.Resolve(selectorTemplate, secret)
	}

	propertyPattern := viper.GetString(env.SecretFilePropertyPattern)
	if len(propertyPattern) < 1 {
		// use root level
		if content != "" {
			// use last path segment of selector
			if len(selectorTemplate) < 1 {
				// illegal configuration, should never happen
				logrus.Warnf("Single value but no selector found for %s/%s", secret.Namespace, secret.Name)
				return make(map[interface{}]interface{})
			}
			key := selectorTemplate
			if strings.Contains(selectorTemplate, ".") {
				// use last path segment as key
				parts := strings.Split(selectorTemplate, ".")
				key = parts[len(parts)-1]
				// remove tailing braces
				key = strings.Replace(key, "}", "", -1)
			}
			return map[interface{}]interface{}{key: content}
		}
		return mapContent
	}

	propertyPattern = templates.Resolve(propertyPattern, secret)

	result := make(map[interface{}]interface{})
	current := result

	parts := strings.Split(propertyPattern, ".")
	for i, s := range parts {
		if i < len(parts)-1 {
			// still need to put maps...
			childMap := make(map[interface{}]interface{})
			current[s] = childMap
			current = childMap
		} else {
			// finally reached child
			if content != "" {
				current[s] = content
			} else {
				current[s] = mapContent
			}
		}
	}
	return result
}
