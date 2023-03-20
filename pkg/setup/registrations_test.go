package setup

import (
	"testing"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestCreateSecretSelectFilter_simple(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// equals
	filter := createSecretSelectFilter("foo=bar")

	g.Expect(filter.Create(buildLabelsEvent(map[string]string{
		"foo": "bar",
	}))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildLabelsEvent(map[string]string{
		"foo": "baz",
	}))).To(gomega.BeFalse())
}

func TestCreateSecretSelectFilter_in(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	filter := createSecretSelectFilter("foo in (bar, baz)")

	g.Expect(filter.Create(buildLabelsEvent(map[string]string{
		"foo": "bar",
	}))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildLabelsEvent(map[string]string{
		"foo": "baz",
	}))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildLabelsEvent(map[string]string{
		"foo": "42",
	}))).To(gomega.BeFalse())
}

func TestCreateNameSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// without namespace matcher -> only check for name
	filter := createNameSelector("^foo-bar-.*$")
	g.Expect(filter.Create(buildNameEvent("foo-bar-"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNameEvent("foo-bar-1"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNameEvent("foo-bar----_"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNameEvent("foo-bar"))).To(gomega.BeFalse())
	g.Expect(filter.Create(buildNameEvent("ffoo-bar-"))).To(gomega.BeFalse())

	// add namespace matcher
	defer viper.Reset()

	// match all
	viper.Set(env.SecretNamespaceSelector, "")
	filter = createNameSelector("^foo-bar-.*$")
	g.Expect(filter.Create(buildNameEvent("foo-bar-"))).To(gomega.BeTrue())

	// match in list
	viper.Set(env.SecretNamespaceSelector, "stuff,what,ever")
	filter = createNameSelector("^foo-bar-.*$")
	g.Expect(filter.Create(buildNameEvent("foo-bar-"))).To(gomega.BeTrue())

	// no match in list
	viper.Set(env.SecretNamespaceSelector, "what,ever")
	filter = createNameSelector("^foo-bar-.*$")
	g.Expect(filter.Create(buildNameEvent("foo-bar-"))).To(gomega.BeFalse())
}

func TestCreateNameSelector_invalid(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Log("Expected panic to happen")
			t.FailNow()
		}
	}()
	createNameSelector("[")
}

func buildNameEvent(name string) event.CreateEvent {
	return event.CreateEvent{
		Object: &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "stuff",
			},
		},
	}
}

func buildLabelsEvent(labels map[string]string) event.CreateEvent {
	return event.CreateEvent{
		Object: &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
		},
	}
}
