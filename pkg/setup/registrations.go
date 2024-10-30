package setup

import (
	"errors"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jaconi-io/secret-file-provider/pkg/controllers/secrets"
	"github.com/jaconi-io/secret-file-provider/pkg/env"

	"github.com/spf13/viper"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func RegisterControllers(mgr manager.Manager) error {
	filter, err := createFilter()
	if err != nil {
		return err
	}

	slog.Info("registering secret controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(filter).
		Complete(&secrets.Reconciler{Client: mgr.GetClient()})
}

// createFilter creates secret read filters based on either name / namespace or label selector.
func createFilter() (predicate.Predicate, error) {
	labelSelector := viper.GetString(env.SecretLabelSelector)
	nameSelector := viper.GetString(env.SecretNameSelector)
	namespaceSelector := strings.Split(viper.GetString(env.SecretNamespaceSelector), ",")

	if labelSelector != "" && nameSelector != "" {
		return nil, errors.New("name and label selector are set")
	}

	if labelSelector != "" {
		labelSelectorPredicate, err := matchByLabelSelector(labelSelector)
		if err != nil {
			return nil, err
		}

		namespacePredicate := matchByNamespace(namespaceSelector)

		return predicate.And(matchRelevantEvents(), namespacePredicate, labelSelectorPredicate), nil
	}

	if nameSelector != "" {
		namePredicate, err := matchByName(nameSelector)
		if err != nil {
			return nil, err
		}

		namespacePredicate := matchByNamespace(namespaceSelector)

		return predicate.And(matchRelevantEvents(), namePredicate, namespacePredicate), nil
	}

	return nil, errors.New("no secret selector set")
}

// matchByLabelSelector returns a predicate matching objects by a Kubernetes label selector.
func matchByLabelSelector(selectFilter string) (predicate.Predicate, error) {
	selector, err := metav1.ParseToLabelSelector(selectFilter)
	if err != nil {
		return nil, err
	}

	return predicate.LabelSelectorPredicate(*selector)
}

// matchByName returns a predicate matching objects name, using a regular expression.
func matchByName(regexString string) (predicate.Predicate, error) {
	regex, err := regexp.CompilePOSIX(regexString)
	if err != nil {
		return nil, err
	}

	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		return regex.Match([]byte(object.GetName()))
	}), nil
}

// matchByNamespace returns a predicate matching an object by its namespace. If the list of namespaces is empty, match
// all objects.
func matchByNamespace(namespaces []string) predicate.Predicate {
	if len(namespaces) == 0 {
		return predicate.NewPredicateFuncs(func(object client.Object) bool {
			return true
		})
	}

	namespaceMap := map[string]struct{}{}
	for _, namespace := range namespaces {
		namespaceMap[namespace] = struct{}{}
	}

	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		_, ok := namespaceMap[object.GetNamespace()]
		return ok
	})
}

// matchRelevantEvents returns a predicate matching all objects for the relevant events create, update, and (optionally)
// delete.
func matchRelevantEvents() predicate.Predicate {
	funcs := predicate.Funcs{
		CreateFunc: func(_ event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(_ event.DeleteEvent) bool {
			return viper.GetBool(env.SecretDeletionWatch)
		},
		GenericFunc: func(_ event.GenericEvent) bool {
			return false
		},
		UpdateFunc: func(_ event.UpdateEvent) bool {
			return true
		},
	}

	return funcs
}
