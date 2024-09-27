package templates

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
)

// Resolve resolves a given go template pattern with the content of the given secret.
func Resolve(pattern string, secret *corev1.Secret) string {

	if !strings.Contains(pattern, "{{") {
		// no template involved, return as is
		return pattern
	}

	// Handle special '.Data' case for secrets, where content is stored in binary format
	patternToApply := pattern
	if strings.Contains(pattern, ".Data") {
		// copy binary .Data secrets to .StringData secrets
		if secret.StringData == nil {
			secret.StringData = map[string]string{}
		}
		for k, v := range secret.Data {
			secret.StringData[k] = string(v)
		}
		// replace accessor in pattern
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
		slog.Error("failed to parse template", "template", pattern, "error", err)
		os.Exit(1)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, secret)
	if err != nil {
		slog.Error("failed to execute template", "namespace", secret.Namespace, "name", secret.Name, "error", err)
		return ""
	}
	return string(buf.Bytes())
}
