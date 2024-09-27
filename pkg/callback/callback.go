package callback

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/logger"
	"github.com/jaconi-io/secret-file-provider/pkg/templates"

	"github.com/spf13/viper"

	corev1 "k8s.io/api/core/v1"
)

// Call a pre-configured HTTP callback endpoint to notify about changes to the given secret. Returns an error, if the
// HTTP call fails or if it returns a non-2xx status code. The boolean indicates if a retry might solve the issue. If
// the error is nil, the boolean has no meaning.
func Call(secret *corev1.Secret) (bool, error) {
	callbackURL := viper.GetString(env.CallbackURL)
	if callbackURL == "" {
		logger.New(secret).Debug("No callback URL has been configured. Skipping callback.")
		return false, nil
	}

	method := viper.GetString(env.CallbackMethod)

	var bodyReader io.Reader
	switch method {
	case http.MethodPatch, http.MethodPost, http.MethodPut:
		bodyReader = body(secret)
	case http.MethodDelete, http.MethodGet, http.MethodHead:
		bodyReader = nil
	default:
		return false, fmt.Errorf("unsupported HTTP method (%s) for callback", method)
	}

	req, err := http.NewRequest(method, callbackURL, bodyReader)
	if err != nil {
		return false, fmt.Errorf("could not create callback request: %w", err)
	}

	req.Header.Add("Content-Type", viper.GetString(env.CallbackContentType))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return true, fmt.Errorf("error during callback request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 405 {
		return false, fmt.Errorf("HTTP method (%s) is not supported by the server", method)
	}

	if resp.StatusCode > 299 {
		return true, fmt.Errorf("callback returned unexpected status code %d", resp.StatusCode)
	}

	return false, nil
}

func body(secret *corev1.Secret) io.Reader {
	body := viper.GetString(env.CallbackBody)
	if body == "" {
		return strings.NewReader("")
	}
	return strings.NewReader(templates.Resolve(body, secret))
}
