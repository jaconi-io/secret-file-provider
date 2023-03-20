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
			log.Fatalf("Unsupported HTTP method %s for callback", method)
		}
	}
	if err != nil {
		return err
	}
	if resp.StatusCode == 405 {
		// method not allowed, fail fast
		log.Panicf("HTTP method %s not supported by server", method)
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("Unexpected status code %d", resp.StatusCode)
	}
	log.Debugf("Successfuly ran callback %s %s", method, callbackUrl)
	return nil
}

func body(secret *corev1.Secret) io.Reader {
	body := viper.GetString(env.CallbackBody)
	if body == "" {
		return strings.NewReader("")
	}
	return strings.NewReader(templates.Resolve(body, secret))
}
