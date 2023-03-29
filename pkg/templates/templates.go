package templates

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// Resolve resolves a given go template pattern with the content of the given secret.
func Resolve(pattern string, secret *corev1.Secret) string {

	if !strings.Contains(pattern, "{{") {
		// no template involved, return as is
		return pattern
	}

	// Handle special '.Data' case for secrets, where content is stored in binary format:
	patternToApply := pattern
	if strings.Contains(pattern, ".Data") {
		// move to stringsecrets
		if secret.StringData == nil {
			secret.StringData = map[string]string{}
		}
		for k, v := range secret.Data {
			secret.StringData[k] = string(v)
		}
		patternToApply = strings.Replace(patternToApply, ".Data", ".StringData", -1)
	}

	// TODO do we want to support extra functions?
	funcMap := template.FuncMap{
		"split":  strings.Split,
		"splitN": strings.SplitN,
	}
	// see https://pkg.go.dev/text/template
	tmpl, err := template.New("test").Funcs(funcMap).Parse(patternToApply)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to parse template '%s'", pattern)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, secret)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to execute template for %s/%s", secret.Namespace, secret.Name)
		return ""
	}
	return string(buf.Bytes())
}
