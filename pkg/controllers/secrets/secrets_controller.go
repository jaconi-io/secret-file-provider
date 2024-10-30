package secrets

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jaconi-io/secret-file-provider/pkg/callback"
	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/file"
	"github.com/jaconi-io/secret-file-provider/pkg/logger"
	"github.com/jaconi-io/secret-file-provider/pkg/maps"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
}

var _ reconcile.Reconciler = &Reconciler{}

func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, request.NamespacedName, secret); err != nil {
		if errors.IsNotFound(err) {
			// do nothing
			return reconcile.Result{}, nil
		}
		slog.Error("failed to read secret", "error", err)
		return reconcile.Result{}, err
	}

	if secret.DeletionTimestamp != nil {
		if !viper.GetBool(env.SecretDeletionWatch) {
			// ignore deletion
			return reconcile.Result{}, nil
		}
		err := change(secret, remove)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Remove the finalizer, once the cleanup completed successfully.
		if _, err := controllerutil.CreateOrPatch(ctx, r.Client, secret, func() error {
			controllerutil.RemoveFinalizer(secret, env.GetFinalizer())
			return nil
		}); err != nil {
			return reconcile.Result{}, fmt.Errorf("removing finalizer failed: %w", err)
		}

		return reconcile.Result{}, nil
	}

	if viper.GetBool(env.SecretDeletionWatch) {
		// Add a finalizer to ensure proper cleanup.
		if _, err := controllerutil.CreateOrPatch(ctx, r.Client, secret, func() error {
			controllerutil.AddFinalizer(secret, env.GetFinalizer())
			return nil
		}); err != nil {
			return reconcile.Result{}, fmt.Errorf("adding finalizer failed: %w", err)
		}
	}

	return reconcile.Result{}, change(secret, add)
}

// change will call the given change function on the secret and call a probably
// existing callback endpoint
// Returns an error if anything went wrong
func change(secret *corev1.Secret, changeFunc func(*corev1.Secret) error) error {
	err := changeFunc(secret)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}
	retry, err := callback.Call(secret)
	if err != nil {
		if retry {
			return fmt.Errorf("failed to run callback: %w", err)
		} else {
			logger.New(secret).Error("failed to run callback", "error", err)
			os.Exit(1)
		}
	}
	return nil
}

// remove will remove the files or file content, belonging to the given secret
// Returns potential error
func remove(secret *corev1.Secret) error {
	log := logger.New(secret)
	log.Debug("Removing content for secret")

	// 1. read existing file content
	f, err := file.Name(secret)
	if err != nil {
		return err
	}
	existingContent := file.ReadAll(log, f)

	// 2. read content from secret
	newContent, err := readSecretContent(secret)
	if err != nil {
		return err
	}

	// 3. drop new entries from existing map
	resultingMap := maps.Drop(existingContent, newContent)

	// 4. write to file
	return file.WriteAll(log, f, resultingMap)
}

// add will create the files or file content, belonging to the given secret
// Returns potential error
func add(secret *corev1.Secret) error {
	log := logger.New(secret)
	log.Debug("Adding content for secret")

	// 1. read existing file content
	f, err := file.Name(secret)
	if err != nil {
		return err
	}
	existingContent := file.ReadAll(log, f)

	// 2. read content from secret
	newContent, err := readSecretContent(secret)
	if err != nil {
		return err
	}

	// 3. merge maps
	resultingMap := maps.Union(existingContent, newContent)

	// 4. write to file
	return file.WriteAll(log, f, resultingMap)
}
