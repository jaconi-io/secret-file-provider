package secrets

import (
	"context"

	"github.com/jaconi-io/secret-file-provider/pkg/callback"
	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/file"
	"github.com/jaconi-io/secret-file-provider/pkg/logger"
	"github.com/jaconi-io/secret-file-provider/pkg/maps"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	Client client.Client
}

var _ reconcile.Reconciler = &Reconciler{}

func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, request.NamespacedName, secret); err != nil {
		if errors.IsNotFound(err) {
			// do nothing
			return reconcile.Result{}, nil
		}
		logrus.WithError(err).Error("Failed to read secret")
		return reconcile.Result{}, err
	}

	if secret.DeletionTimestamp != nil {
		// TODO this might be problematic and can only be overcome with finalizers
		return reconcile.Result{}, change(secret, remove)
	}

	return reconcile.Result{}, change(secret, add)
}

func change(secret *corev1.Secret, changeFunc func(*corev1.Secret) error) error {
	log := logger.New(secret)
	err := changeFunc(secret)
	if err != nil {
		return err
	}
	if err != nil {
		log.WithError(err).Error("Failed to update content")
		return err
	}
	err = callback.Call(secret)
	if err != nil {
		log.WithError(err).Error("Failed to run callback")
		return err
	}
	return nil
}

func remove(secret *corev1.Secret) error {
	log := logger.New(secret)
	log.Debug("Removing content for secret")

	// 1. read existing file content
	f := file.Name(secret)
	existingContent := file.ReadAll(log, f)

	// 2. read content from secret
	newContent := readSecretContent(secret)

	// 3. convert keys if necessary
	convertedKeyMap := maps.TransformKeys(newContent, viper.GetString(env.SecretKeyTransformation))

	// 4. drop new entries from existing map
	resultingMap := maps.Drop(existingContent, convertedKeyMap)

	// 5. write to file
	return file.WriteAll(log, f, resultingMap)
}

func add(secret *corev1.Secret) error {
	log := logger.New(secret)
	log.Debug("Adding content for secret")

	// 1. read existing file content
	f := file.Name(secret)
	existingContent := file.ReadAll(log, f)

	// 2. read content from secret
	newContent := readSecretContent(secret)

	// 3. convert keys if necessary
	convertedKeyMap := maps.TransformKeys(newContent, viper.GetString(env.SecretKeyTransformation))

	// 4. merge maps
	resultingMap := maps.Union(existingContent, convertedKeyMap)

	// 5. write to file
	return file.WriteAll(log, f, resultingMap)
}
