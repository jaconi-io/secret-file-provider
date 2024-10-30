package templates

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
)

// TODO: Do we want to support additional functions?
var funcMap = template.FuncMap{
	"split":  strings.Split,
	"splitN": strings.SplitN,
}

// Render a given Go template with the content of the given Kubernetes secret.
func Render(pattern string, secret *corev1.Secret) (string, error) {
	if !strings.Contains(pattern, "{{") {
		// Not a Go template. Return as is.
		return pattern, nil
	}

	// Copy binary '.Data' to '.StringData'.
	if strings.Contains(pattern, ".Data") {
		if secret.StringData == nil {
			secret.StringData = map[string]string{}
		}

		for k, v := range secret.Data {
			secret.StringData[k] = string(v)
		}

		// Replace occurrences in pattern.
		pattern = strings.Replace(pattern, ".Data", ".StringData", -1)
	}

	// See https://pkg.go.dev/text/template
	tmpl, err := template.New("").Funcs(funcMap).Parse(pattern)
	if err != nil {
		return "", fmt.Errorf("parsing template %q failed: %w", pattern, err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, secret)
	if err != nil {
		return "", fmt.Errorf("executing template %q with secret %s/%s failed: %w", pattern, secret.Namespace, secret.Name, err)
	}

	return buf.String(), nil
}
