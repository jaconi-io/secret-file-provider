package e2e_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestExample(t *testing.T) {
	g := NewGomegaWithT(t)

	feat := features.New("test").
		WithSetup("do stuff", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			client, err := c.NewClient()
			g.Expect(err).NotTo(HaveOccurred())

			deployment := deployment()
			for _, v := range []k8s.Object{deployment, role(), sa(), roleBinding(), secret()} {
				err := client.Resources().Create(ctx, v)
				g.Expect(err).NotTo(HaveOccurred())
			}

			err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
			g.Expect(err).NotTo(HaveOccurred())

			return ctx
		}).
		Assess("check stuff", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			client, err := c.NewClient()
			g.Expect(err).NotTo(HaveOccurred())

			pods := &corev1.PodList{}
			err = client.Resources(c.Namespace()).List(ctx, pods)
			g.Expect(err).NotTo(HaveOccurred())

			var stdout, stderr bytes.Buffer
			podName := pods.Items[0].Name
			command := []string{"cat", "/secrets/test.yaml"}

			err = client.Resources().ExecInPod(ctx, c.Namespace(), podName, "helper", command, &stdout, &stderr)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(stdout.String()).To(Equal("foo: bar\n"))
			g.Expect(stderr.String()).To(Equal(""))
			return ctx
		}).Feature()

	_ = testenv.Test(t, feat)
}

func deployment() *appsv1.Deployment {
	labels := map[string]string{"app.kubernetes.io/name": "secret-file-provider"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e",
			Namespace: "e2e",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "helper",
							Image:   "busybox",
							Command: []string{"sleep", "infinity"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrets",
									ReadOnly:  true,
									MountPath: "/secrets",
								},
							},
						},
						{
							Name:  "secret-file-provider",
							Image: "ghcr.io/jaconi-io/secret-file-provider:latest",
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "SECRET_FILE_NAME_PATTERN",
									Value: "/secrets/test.yaml",
								},
								{
									Name:  "SECRET_SELECTOR_NAME",
									Value: ".*",
								},
								{
									Name:  "SECRET_SELECTOR_NAMESPACE",
									Value: "e2e",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "secrets",
									MountPath: "/secrets",
								},
							},
						},
					},
					ServiceAccountName: "secret-file-provider",
					Volumes: []corev1.Volume{
						{
							Name: "secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
				},
			},
		},
	}
}

func secret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "e2e",
		},
		StringData: map[string]string{
			"foo": "bar",
		},
	}
}

func role() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-file-provider",
			Namespace: "e2e",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"list", "patch", "watch"},
			},
		},
	}
}

func roleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-file-provider",
			Namespace: "e2e",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "secret-file-provider",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "secret-file-provider",
				Namespace: "e2e",
			},
		},
	}
}

func sa() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-file-provider",
			Namespace: "e2e",
		},
	}
}
