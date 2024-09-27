package e2e_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/jaconi-io/secret-file-provider/e2e/testdata"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
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
		}).
		Assess("logs are written", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			pods := &corev1.PodList{}
			err := c.Client().Resources(c.Namespace()).List(ctx, pods)
			g.Expect(err).NotTo(HaveOccurred())

			pod := pods.Items[0]

			logs, err := readLogs(ctx, c, pod)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(logs).To(ContainSubstring("successfuly written"))
			g.Expect(logs).To(ContainSubstring("path=/secrets/secrets.yaml"))
			g.Expect(logs).To(ContainSubstring("name=example"))
			g.Expect(logs).To(ContainSubstring("namespace=" + pod.Namespace))

			return ctx
		}).Feature()

	_ = testenv.Test(t, feat)
}

// readLogs gets the log output (stdout + stderr) from 'secret-file-provider' in a given pod.
func readLogs(ctx context.Context, c *envconf.Config, pod corev1.Pod) (string, error) {

	podLogOpts := &corev1.PodLogOptions{Container: "secret-file-provider"}
	clientset, err := kubernetes.NewForConfig(c.Client().RESTConfig())
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
