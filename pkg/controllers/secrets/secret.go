package secrets

import (
	"log/slog"
	"strings"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/templates"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

// readSecretContent reads the content of the given secret as key value pairs into a map. Note that
// this will also apply the secret content selector, so not every key of the original secret might
// be represented in the resulting map.
// Note that this also applies the propertyPattern, which means that this might nest the actual secrets
// inside other maps.
//
// Example:
//
//	secret.selector.content     = "{{.Data.CLIENT_ID}}"
//	secret.file.property.pattern="foo.bar.clientIds.{{.ObjectMeta.Labels.company}}"
//
// The resulting map will be:
//
//	foo: {
//	  bar: {
//		clientIds: {
//		  acme: the-acme-id,
//	      foobar: some-foobar-client,
//	      ...
//	    }
//	  }
//	}
func readSecretContent(secret *corev1.Secret) map[interface{}]interface{} {
	propertyPattern := viper.GetString(env.SecretFilePropertyPattern)

	mapContent, stringContent := extractContent(secret)

	// if no additional properties: put content into plain map
	if len(propertyPattern) < 1 {
		if stringContent == "" {
			// no nesting required, no plain content -> we are done at this point and
			// can return the already read in map
			return mapContent
		}
		return processSingleElement(stringContent)
	}

	return nestAdditionalProperties(secret, mapContent, stringContent)
}

// extractContent stores content information either as map or as plain string, depending on selector
func extractContent(secret *corev1.Secret) (map[interface{}]interface{}, string) {
	selectorTemplate := viper.GetString(env.SecretContentSelector)

	mapContent := make(map[interface{}]interface{})
	stringContent := ""

	// fill content with secret data and selectorTemplate
	if len(selectorTemplate) < 1 {
		// put all into map
		for k, v := range secret.Data {
			mapContent[transform(k)] = string(v)
		}
	} else if !strings.Contains(selectorTemplate, "{{") {
		// not a go template, log warning and put all into map
		slog.Warn("illegal selector pattern; expecting go template", "pattern", selectorTemplate)
		for k, v := range secret.Data {
			mapContent[transform(k)] = string(v)
		}
	} else {
		// resolve template to string; do not put into map, as this is intended to be a plain string
		stringContent = templates.Resolve(selectorTemplate, secret)
	}
	return mapContent, stringContent
}

// processSingleElement creates a map containing only the given string value as value and the last
// path segment of the content selector as key
func processSingleElement(stringContent string) map[interface{}]interface{} {
	selectorTemplate := viper.GetString(env.SecretContentSelector)

	if len(selectorTemplate) < 1 {
		// illegal configuration, should never happen
		slog.Warn("single value but no selector found")
		return make(map[interface{}]interface{})
	}

	// use last path segment of selector as key for the new map
	key := selectorTemplate
	if strings.Contains(selectorTemplate, ".") {
		parts := strings.Split(selectorTemplate, ".")
		key = parts[len(parts)-1]
		// remove tailing braces
		key = strings.Replace(key, "}", "", -1)
		// make sure the key (refering to secret key) is transformed if necessary
		key = transform(key)
	}
	return map[interface{}]interface{}{key: stringContent}
}

// nestAdditionalProperties will attach either the given map- or string-content to a mandatory property pattern
// prefix, gotten via [env.SecretFilePropertyPattern].
func nestAdditionalProperties(secret *corev1.Secret, mapContent map[interface{}]interface{}, stringContent string) map[interface{}]interface{} {
	propertyPattern := viper.GetString(env.SecretFilePropertyPattern)
	propertyPath := templates.Resolve(propertyPattern, secret)

	result := make(map[interface{}]interface{})
	current := result

	// for each part in the property path: create a nested child map
	properties := strings.Split(propertyPath, ".")
	for idx, prop := range properties {
		if idx < len(properties)-1 {
			// still need to nest maps...
			childMap := make(map[interface{}]interface{})
			current[prop] = childMap
			current = childMap
		} else {
			// finally reached leaf
			if stringContent != "" {
				current[prop] = stringContent
			} else {
				current[prop] = mapContent
			}
		}
	}
	return result
}

func transform(key string) string {
	transform := viper.GetString(env.SecretKeyTransformation)
	if function, ok := keyTransformFunctions[transform]; ok {
		return function(key)
	}
	return key
}
