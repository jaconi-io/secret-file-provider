package countingfinalizer_test

import (
	"testing"

	. "github.com/jaconi-io/secret-file-provider/pkg/countingfinalizer"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateFinalizer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	secret := &corev1.Secret{}

	Increment(secret, "test")
	g.Expect(secret.Finalizers).To(gomega.HaveExactElements("test1"))
}

func TestIncrementFinalizer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Finalizers: []string{"test1"},
		},
	}

	Increment(secret, "test")
	g.Expect(secret.Finalizers).To(gomega.HaveExactElements("test2"))
}

func TestIncrementInvalidFinalizer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Finalizers: []string{"testfoo"},
		},
	}

	Increment(secret, "test")
	g.Expect(secret.Finalizers).To(gomega.HaveExactElements("testfoo", "test1"))
}

func TestDecrementMissingFinalizer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	secret := &corev1.Secret{}

	Decrement(secret, "test")
	g.Expect(secret.Finalizers).To(gomega.BeEmpty())
}

func TestRemoveFinalizer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Finalizers: []string{"test1"},
		},
	}

	Decrement(secret, "test")
	g.Expect(secret.Finalizers).To(gomega.BeEmpty())
}

func TestDecrementFinalizer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Finalizers: []string{"test2"},
		},
	}

	Decrement(secret, "test")
	g.Expect(secret.Finalizers).To(gomega.HaveExactElements("test1"))
}
