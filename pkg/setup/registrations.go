package setup

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/jaconi-io/secret-file-provider/pkg/controllers/secrets"
	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func RegisterControllers(mgr manager.Manager) {
	slog.Info("register reconcilers...")
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(createFilter()).
		Complete(&secrets.Reconciler{Client: mgr.GetClient()}); err != nil {
		slog.Error("could not create controller", "error", err)
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
	slog.Error("no secret selector set")
	os.Exit(1)
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
	return createFunctionSelectFilter(namespaces, regexString)
}

// createSecretSelectFilter creates a predicate to check K8s label selector matches for reconciles.
func createSecretSelectFilter(selectFilter string) predicate.Predicate {

	// TODO namespace!
	// TODO use real k8s selector instead of object meta selector?
	selector, err := metav1.ParseToLabelSelector(selectFilter)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to parse given selector %s", selectFilter), "error", err)
		os.Exit(1)
	}
	filter, err := predicate.LabelSelectorPredicate(*selector)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to convert selector %s", selectFilter), "error", err)
		os.Exit(1)
	}
	return predicate.And(filter, createAllFunctionSelectFilter())
}

func createFunctionSelectFilter(namespaces map[string]struct{}, regexString string) predicate.Predicate {
	regex := regexp.MustCompilePOSIX(regexString)
	funcs := predicate.Funcs{
		CreateFunc: func(ce event.CreateEvent) bool {
			return namespaceMatch(namespaces, ce.Object.GetNamespace()) && regex.Match([]byte(ce.Object.GetName()))
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			return namespaceMatch(namespaces, ue.ObjectNew.GetNamespace()) && regex.Match([]byte(ue.ObjectNew.GetName()))
		},
	}
	if viper.GetBool(env.SecretDeletionWatch) {
		funcs.DeleteFunc = func(de event.DeleteEvent) bool {
			return namespaceMatch(namespaces, de.Object.GetNamespace()) && regex.Match([]byte(de.Object.GetName()))
		}
	}
	return funcs
}

func createAllFunctionSelectFilter() predicate.Predicate {
	funcs := predicate.Funcs{
		CreateFunc: func(_ event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(_ event.UpdateEvent) bool {
			return true
		},
	}
	if viper.GetBool(env.SecretDeletionWatch) {
		funcs.DeleteFunc = func(_ event.DeleteEvent) bool {
			return true
		}
	}
	return funcs
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
