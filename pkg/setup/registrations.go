package setup

import (
	"regexp"
	"strings"

	"github.com/jaconi-io/secret-file-provider/pkg/controllers/secrets"
	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func RegisterControllers(mgr manager.Manager) {

	logrus.Info("Register Reconcilers...")
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(createFilter()).
		Complete(&secrets.Reconciler{Client: mgr.GetClient()}); err != nil {
		logrus.WithError(err).Fatal("could not create controller")
	}
}

// createFilter creates secret read filters based on either name-/namespace- or label-selector.
func createFilter() predicate.Predicate {
	labelSelector := viper.GetString(env.SecretLabelSelector)
	if len(labelSelector) > 0 {
		return createSecretSelectFilter(labelSelector)
	}
	nameSelector := viper.GetString(env.SecretNameSelector)
	if len(nameSelector) > 0 {
		return createNameSelector(nameSelector)
	}
	logrus.Fatal("No secret selector set")
	return nil
}

// createNameSelector creates a predicate to check for dedicated name-pattern/namespace combinations
// to reconcile on.
func createNameSelector(regexString string) predicate.Predicate {

	namespaces := make(map[string]struct{})
	namespacesString := viper.GetString(env.SecretNamespaceSelector)
	if namespacesString != "" {
		for _, ns := range strings.Split(namespacesString, ",") {
			namespaces[ns] = struct{}{}
		}
	}
	regex := regexp.MustCompilePOSIX(regexString)
	return predicate.Funcs{
		CreateFunc: func(ce event.CreateEvent) bool {
			return namespaceMatch(namespaces, ce.Object.GetNamespace()) && regex.Match([]byte(ce.Object.GetName()))
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			return namespaceMatch(namespaces, ue.ObjectNew.GetNamespace()) && regex.Match([]byte(ue.ObjectNew.GetName()))
		},
		DeleteFunc: func(de event.DeleteEvent) bool {
			return namespaceMatch(namespaces, de.Object.GetNamespace()) && regex.Match([]byte(de.Object.GetName()))
		},
	}
}

func namespaceMatch(namespaces map[string]struct{}, namespace string) bool {
	if len(namespaces) < 1 {
		// match all
		return true
	}
	if _, ok := namespaces[namespace]; ok {
		return true
	}
	return false
}

// createSecretSelectFilter creates a predicate to check K8s label selector matches for reconciles.
func createSecretSelectFilter(selectFilter string) predicate.Predicate {

	// TODO namespace!
	// TODO use real k8s selector instead of object meta selector?
	selector, err := metav1.ParseToLabelSelector(selectFilter)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to parse given selector %s", selectFilter)
	}
	filter, err := predicate.LabelSelectorPredicate(*selector)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed convert selector %s", selectFilter)
	}
	return filter
}
