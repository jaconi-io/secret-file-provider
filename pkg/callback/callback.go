package callback

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/logger"
	"github.com/jaconi-io/secret-file-provider/pkg/templates"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

// Call will run an HTTP call against a preconfigured callback endpoint to notify address about changes
// written for the given secret.
// Returns an error if HTTP call fails or if it retrieves a non-2xx status response.
func Call(secret *corev1.Secret) error {
	log := logger.New(secret)
	callbackUrl := viper.GetString(env.CallbackUrl)
	if callbackUrl == "" {
		log.Debug("No callback URL set. Skip operation")
		return nil
	}

	method := viper.GetString(env.CallbackMethod)
	var resp *http.Response
	var err error

	switch method {
	case "GET":
		{
			resp, err = http.Get(callbackUrl)
			break
		}
	case "POST":
		{
			resp, err = http.Post(callbackUrl, viper.GetString(env.CallbackContenttype), body(secret))
			break
		}
	case "HEAD":
		{
			resp, err = http.Head(callbackUrl)
			break
		}
	case "PATCH":
	case "PUT":
		{
			req, errx := http.NewRequest(method, callbackUrl, body(secret))
			if errx != nil {
				return errx
			}
			req.Header.Add("Content-Type", viper.GetString(env.CallbackContenttype))
			resp, err = http.DefaultClient.Do(req)
			break
		}
	case "DELETE":
		{
			req, errx := http.NewRequest(method, callbackUrl, nil)
			if errx != nil {
				return errx
			}
			resp, err = http.DefaultClient.Do(req)
			break
		}
	default:
		{
			// not supported, fail fast
			log.Error("unsupported HTTP method for callback", "method", method)
			os.Exit(1)
		}
	}
	if err != nil {
		return err
	}
	if resp.StatusCode == 405 {
		// method not allowed, fail fast
		log.Error("HTTP method not supported by server", "method", method)
		panic(fmt.Errorf("HTTP method %s not supported by server", method))
	}
	if resp.StatusCode > 299 {
		// might be a temporary issue, do reconcile
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed to call callback URL", "url", callbackUrl, "error", err)
		os.Exit(1)
	}
	log.Debug("successfully ran callback", "method", method, "url", callbackUrl, "response", string(bodyBytes))
	return nil
}

func body(secret *corev1.Secret) io.Reader {
	body := viper.GetString(env.CallbackBody)
	if body == "" {
		return strings.NewReader("")
	}
	return strings.NewReader(templates.Resolve(body, secret))
}
