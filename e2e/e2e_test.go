package e2e_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/jaconi-io/secret-file-provider/e2e/testdata"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestExample(t *testing.T) {
	g := NewGomegaWithT(t)

	feat := features.New("Basic functionality").
		WithSetup("testdata", testdata.Setup).
		Assess("secret file is created", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			pods := &corev1.PodList{}
			err := c.Client().Resources(c.Namespace()).List(ctx, pods)
			g.Expect(err).NotTo(HaveOccurred())

			var stdout, stderr bytes.Buffer
			command := []string{"cat", "/secrets/secrets.yaml"}
			err = c.Client().Resources().ExecInPod(ctx, c.Namespace(), pods.Items[0].Name, "helper", command, &stdout, &stderr)

			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(stdout.String()).To(Equal("foo: bar\n"))
			g.Expect(stderr.String()).To(Equal(""))

			return ctx
		}).Feature()

	_ = testenv.Test(t, feat)
}
