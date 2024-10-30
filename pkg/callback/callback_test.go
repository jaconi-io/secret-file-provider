package callback

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCallEmptyURL(t *testing.T) {
	g := NewGomegaWithT(t)

	retry, err := Call(&corev1.Secret{})

	g.Expect(retry).To(BeFalse())
	g.Expect(err).To(BeNil())
}

func TestCallInvalidMethod(t *testing.T) {
	g := NewGomegaWithT(t)
	defer viper.Reset()

	viper.Set(env.CallbackMethod, http.MethodOptions)
	viper.Set(env.CallbackURL, "http://localhost/callback")

	retry, err := Call(&corev1.Secret{})

	g.Expect(retry).To(BeFalse())
	g.Expect(err).To(MatchError("unsupported HTTP method (OPTIONS) for callback"))
}

func TestCallBrokenURL(t *testing.T) {
	g := NewGomegaWithT(t)
	defer viper.Reset()

	viper.Set(env.CallbackMethod, http.MethodGet)
	viper.Set(env.CallbackURL, "\n")

	retry, err := Call(&corev1.Secret{})

	g.Expect(retry).To(BeFalse())
	g.Expect(err).To(MatchError("could not create callback request: parse \"\\n\": net/url: invalid control character in URL"))
}

func TestCallInvalidURL(t *testing.T) {
	g := NewGomegaWithT(t)
	defer viper.Reset()

	viper.Set(env.CallbackMethod, http.MethodGet)
	viper.Set(env.CallbackURL, "invalid")

	retry, err := Call(&corev1.Secret{})

	g.Expect(retry).To(BeTrue()) // Might be a timeout.
	g.Expect(err).To(MatchError("error during callback request: Get \"invalid\": unsupported protocol scheme \"\""))
}

func TestCallGet(t *testing.T) {
	defer viper.Reset()

	for _, tt := range []struct {
		StatusCode int
		Retry      bool
		Error      string
	}{
		{400, true, "callback returned unexpected status code 400"},
		{405, false, "HTTP method (GET) is not supported by the server"},
		{500, true, "callback returned unexpected status code 500"},
	} {
		t.Run(strconv.Itoa(tt.StatusCode), func(t *testing.T) {
			g := NewGomegaWithT(t)

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/callback" {
					rw.WriteHeader(tt.StatusCode)
				}
			}))
			defer server.Close()

			viper.Set(env.CallbackMethod, http.MethodGet)
			viper.Set(env.CallbackURL, server.URL+"/callback")

			retry, err := Call(&corev1.Secret{})

			g.Expect(retry).To(Equal(tt.Retry))
			g.Expect(err).To(MatchError(tt.Error))
		})
	}
}

func TestCallPostEmptyBody(t *testing.T) {
	g := NewGomegaWithT(t)
	defer viper.Reset()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/callback" {
			g.Expect(req.ContentLength).To(BeZero())
		}
	}))
	defer server.Close()

	viper.Set(env.CallbackMethod, http.MethodPost)
	viper.Set(env.CallbackURL, server.URL+"/callback")

	retry, err := Call(&corev1.Secret{})

	g.Expect(retry).To(BeFalse())
	g.Expect(err).To(BeNil())
}

func TestCallPostInvalidBody(t *testing.T) {
	g := NewGomegaWithT(t)
	defer viper.Reset()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/callback" {
			g.Expect(req.ContentLength).To(BeNumerically(">", 0))
			bytes := make([]byte, req.ContentLength)
			req.Body.Read(bytes)
			g.Expect(string(bytes)).To(Equal(`{"updated":"foo"}`))
		}
	}))
	defer server.Close()

	// The body template is tested within the server implementation!
	viper.Set(env.CallbackBody, `{"updated":"{{.ObjectMeta.Name}"}`)
	viper.Set(env.CallbackMethod, http.MethodPost)
	viper.Set(env.CallbackURL, server.URL+"/callback")

	retry, err := Call(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	})

	g.Expect(retry).To(BeFalse())
	g.Expect(err).To(MatchError("parsing template \"{\\\"updated\\\":\\\"{{.ObjectMeta.Name}\\\"}\" failed: template: :1: bad character U+007D '}'"))
}

func TestCallPost(t *testing.T) {
	g := NewGomegaWithT(t)
	defer viper.Reset()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/callback" {
			g.Expect(req.ContentLength).To(BeNumerically(">", 0))
			bytes := make([]byte, req.ContentLength)
			req.Body.Read(bytes)
			g.Expect(string(bytes)).To(Equal(`{"updated":"foo"}`))
		}
	}))
	defer server.Close()

	// The body template is tested within the server implementation!
	viper.Set(env.CallbackBody, `{"updated":"{{.ObjectMeta.Name}}"}`)
	viper.Set(env.CallbackMethod, http.MethodPost)
	viper.Set(env.CallbackURL, server.URL+"/callback")

	retry, err := Call(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	})

	g.Expect(retry).To(BeFalse())
	g.Expect(err).To(BeNil())
}
