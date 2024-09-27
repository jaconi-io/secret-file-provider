package callback

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCall(t *testing.T) {
	g := NewGomegaWithT(t)
	defer viper.Reset()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	expectedBody := `{"updated":"foo"}`
	viper.Set(env.CallbackBody, `{"updated":"{{.ObjectMeta.Name}}"}`)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if req.URL.Path == "/error400" {
			rw.WriteHeader(400)
			return
		}
		if req.URL.Path == "/error500" {
			rw.WriteHeader(500)
			return
		}
		if req.URL.Path == "/error405" {
			rw.WriteHeader(405)
			return
		}
		if req.URL.Path == "/body" {
			g.Expect(req.ContentLength).Should(BeNumerically(">", 0.0))
			bytes := make([]byte, req.ContentLength)
			req.Body.Read(bytes)
			g.Expect(string(bytes)).To(Equal(expectedBody))
		}
		rw.WriteHeader(200)
		rw.Write([]byte("OK"))
	}))
	defer server.Close()

	// check that body templating works properly -> done within server implementation
	viper.Set(env.CallbackUrl, server.URL+"/body")
	viper.Set(env.CallbackMethod, "POST")

	g.Expect(Call(secret)).To(BeNil())

	// 400 error
	viper.Set(env.CallbackMethod, "GET")
	viper.Set(env.CallbackUrl, server.URL+"/error400")

	err := Call(secret)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(Equal("unexpected status code 400"))

	// 500 error
	viper.Set(env.CallbackUrl, server.URL+"/error500")

	err = Call(secret)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(Equal("unexpected status code 500"))

	// 405 error - method not allowed -> expect panic
	viper.Set(env.CallbackUrl, server.URL+"/error405")
	defer func() {
		if recover() == nil {
			g.Fail("Expect 405 to raise a panic")
		}
	}()
	Call(secret)
}
