package testdata

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func Setup(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
	g := NewGomegaWithT(t)

	err := decoder.DecodeEachFile(
		// Directory is relative to main test!
		ctx, os.DirFS("./testdata"), "*.yaml",
		decoder.CreateHandler(c.Client().Resources()),
		decoder.MutateNamespace(c.Namespace()),
	)
	g.Expect(err).NotTo(HaveOccurred())

	deployments := &appsv1.DeploymentList{}
	err = c.Client().Resources(c.Namespace()).List(ctx, deployments)
	g.Expect(err).NotTo(HaveOccurred())

	err = wait.For(conditions.New(c.Client().Resources()).DeploymentConditionMatch(&deployments.Items[0], appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
	g.Expect(err).NotTo(HaveOccurred())

	return ctx
}
