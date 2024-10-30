package setup

import (
	"testing"

	"github.com/jaconi-io/secret-file-provider/pkg/env"

	"github.com/onsi/gomega"
	"github.com/spf13/viper"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func TestRegisterControllersNoConfiguration(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mgr, err := ctrl.NewManager(&rest.Config{}, manager.Options{})
	g.Expect(err).To(gomega.BeNil())

	err = RegisterControllers(mgr)
	g.Expect(err).To(gomega.MatchError("no secret selector set"))
}

func TestRegisterControllersTooMuchConfiguration(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretLabelSelector, "foo=bar")
	viper.Set(env.SecretNameSelector, ".*")

	mgr, err := ctrl.NewManager(&rest.Config{}, manager.Options{})
	g.Expect(err).To(gomega.BeNil())

	err = RegisterControllers(mgr)
	g.Expect(err).To(gomega.MatchError("name and label selector are set"))
}

func TestRegisterControllers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretLabelSelector, "foo=bar")

	mgr, err := ctrl.NewManager(&rest.Config{}, manager.Options{})
	g.Expect(err).To(gomega.BeNil())

	err = RegisterControllers(mgr)
	g.Expect(err).To(gomega.BeNil())
}

func TestCreateFilterNoConfiguration(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	filter, err := createFilter()
	g.Expect(err).To(gomega.MatchError("no secret selector set"))
	g.Expect(filter).To(gomega.BeNil())
}

func TestCreateFilterTooMuchConfiguration(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretLabelSelector, "foo=bar")
	viper.Set(env.SecretNameSelector, ".*")

	filter, err := createFilter()
	g.Expect(err).To(gomega.MatchError("name and label selector are set"))
	g.Expect(filter).To(gomega.BeNil())
}

func TestCreateFilterLabelSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretLabelSelector, "foo=bar")

	filter, err := createFilter()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(filter.Create(buildLabelsEvent(map[string]string{"foo": "bar"}))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildLabelsEvent(map[string]string{"foo": "baz"}))).To(gomega.BeFalse())
}

func TestCreateFilterInvalidLabelSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretLabelSelector, "foo=bar=42")

	filter, err := createFilter()
	g.Expect(err).To(gomega.MatchError("couldn't parse the selector string \"foo=bar=42\": found '=', expected: ',' or 'end of string'"))
	g.Expect(filter).To(gomega.BeNil())
}

func TestCreateFilterLabelSelectorAndNamespaceSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretLabelSelector, "foo=bar")
	viper.Set(env.SecretNamespaceSelector, "a")

	filter, err := createFilter()
	g.Expect(err).To(gomega.BeNil())

	// Namespace and label match.
	g.Expect(filter.Create(createEvent("a", "", map[string]string{"foo": "bar"}))).To(gomega.BeTrue())

	// Namespace matches but label does not.
	g.Expect(filter.Create(createEvent("a", "", map[string]string{"foo": "baz"}))).To(gomega.BeFalse())

	// Neither namespace nor label match.
	g.Expect(filter.Create(createEvent("b", "", map[string]string{"foo": "baz"}))).To(gomega.BeFalse())
}

func TestCreateFilterNameSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretNameSelector, "foo-.*")

	filter, err := createFilter()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(filter.Create(buildNameEvent("foo-1"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNameEvent("bar-1"))).To(gomega.BeFalse())
}

func TestCreateFilterInvalidNameSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretNameSelector, "[")

	filter, err := createFilter()
	g.Expect(err).To(gomega.MatchError("error parsing regexp: missing closing ]: `[`"))
	g.Expect(filter).To(gomega.BeNil())
}

func TestCreateFilterNameAndNamespaceSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretNameSelector, "foo-.*")
	viper.Set(env.SecretNamespaceSelector, "a")

	filter, err := createFilter()
	g.Expect(err).To(gomega.BeNil())

	// Namespace and name match.
	g.Expect(filter.Create(createEvent("a", "foo-1", map[string]string{}))).To(gomega.BeTrue())

	// Namespace matches but name does not.
	g.Expect(filter.Create(createEvent("a", "bar-1", map[string]string{}))).To(gomega.BeFalse())

	// Neither namespace nor label match.
	g.Expect(filter.Create(createEvent("b", "bar-1", map[string]string{}))).To(gomega.BeFalse())
}

func TestMatchByLabelSelectorSimple(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	filter, err := matchByLabelSelector("foo=bar")
	g.Expect(err).To(gomega.BeNil())

	g.Expect(filter.Create(buildLabelsEvent(map[string]string{
		"foo": "bar",
	}))).To(gomega.BeTrue())

	g.Expect(filter.Create(buildLabelsEvent(map[string]string{
		"foo": "baz",
	}))).To(gomega.BeFalse())
}

func TestMatchByLabelSelectorIn(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	filter, err := matchByLabelSelector("foo in (bar, baz)")
	g.Expect(err).To(gomega.BeNil())

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

func TestMatchByLabelSelectorInvalid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	filter, err := matchByLabelSelector("foo=bar=42")
	g.Expect(err).To(gomega.MatchError("couldn't parse the selector string \"foo=bar=42\": found '=', expected: ',' or 'end of string'"))
	g.Expect(filter).To(gomega.BeNil())
}

func TestMatchByName(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	filter, err := matchByName("^foo-bar-.*$")
	g.Expect(err).To(gomega.BeNil())

	g.Expect(filter.Create(buildNameEvent("foo-bar-"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNameEvent("foo-bar-1"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNameEvent("foo-bar----_"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNameEvent("foo-bar"))).To(gomega.BeFalse())
	g.Expect(filter.Create(buildNameEvent("ffoo-bar-"))).To(gomega.BeFalse())
}

func TestCreateNameSelectorInvalid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	filter, err := matchByName("[")
	g.Expect(err).To(gomega.MatchError("error parsing regexp: missing closing ]: `[`"))
	g.Expect(filter).To(gomega.BeNil())
}

func TestMatchByNamespace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	filter := matchByNamespace([]string{})
	g.Expect(filter.Create(buildNamespaceEvent("foo"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNamespaceEvent("bar"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNamespaceEvent("baz"))).To(gomega.BeTrue())

	filter = matchByNamespace([]string{"foo"})
	g.Expect(filter.Create(buildNamespaceEvent("foo"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNamespaceEvent("bar"))).To(gomega.BeFalse())
	g.Expect(filter.Create(buildNamespaceEvent("baz"))).To(gomega.BeFalse())

	filter = matchByNamespace([]string{"foo", "bar"})
	g.Expect(filter.Create(buildNamespaceEvent("foo"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNamespaceEvent("bar"))).To(gomega.BeTrue())
	g.Expect(filter.Create(buildNamespaceEvent("baz"))).To(gomega.BeFalse())
}

func TestMatchRelevantEvents(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()

	viper.Set(env.SecretDeletionWatch, false)
	filter := matchRelevantEvents()
	g.Expect(filter.Create(event.CreateEvent{})).To(gomega.BeTrue())
	g.Expect(filter.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	g.Expect(filter.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	g.Expect(filter.Update(event.UpdateEvent{})).To(gomega.BeTrue())

	viper.Set(env.SecretDeletionWatch, true)
	filter = matchRelevantEvents()
	g.Expect(filter.Create(event.CreateEvent{})).To(gomega.BeTrue())
	g.Expect(filter.Delete(event.DeleteEvent{})).To(gomega.BeTrue())
	g.Expect(filter.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	g.Expect(filter.Update(event.UpdateEvent{})).To(gomega.BeTrue())
}

func createEvent(namespace string, name string, labels map[string]string) event.CreateEvent {
	return event.CreateEvent{
		Object: &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
				Labels:    labels,
			},
		},
	}
}

func buildLabelsEvent(labels map[string]string) event.CreateEvent {
	return createEvent("", "", labels)
}

func buildNameEvent(name string) event.CreateEvent {
	return createEvent("", name, map[string]string{})
}

func buildNamespaceEvent(namespace string) event.CreateEvent {
	return createEvent(namespace, "", map[string]string{})
}
