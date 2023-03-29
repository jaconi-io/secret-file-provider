package templates

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResolve(t *testing.T) {
	g := NewGomegaWithT(t)

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-super-secret-foo-bar",
			Namespace: "some-namespace",
			Labels: map[string]string{
				"label1": "labelValue1",
			},
			Annotations: map[string]string{
				"annotation1": "annotationValue1",
			},
		},
		Data: map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		},
	}

	// empty string
	g.Expect(Resolve("", &secret)).To(BeEmpty())
	// no pattern
	g.Expect(Resolve("samba-dance", &secret)).To(Equal("samba-dance"))
	// name in pattern
	g.Expect(Resolve("{{.ObjectMeta.Name}}", &secret)).To(Equal("my-super-secret-foo-bar"))
	// label in pattern
	g.Expect(Resolve("{{.ObjectMeta.Labels.label1}}", &secret)).To(Equal("labelValue1"))
	// annotation in pattern
	g.Expect(Resolve("{{.ObjectMeta.Annotations.annotation1}}", &secret)).To(Equal("annotationValue1"))
	// secret value in pattern
	g.Expect(Resolve("{{.Data.key1}}", &secret)).To(Equal("value1"))

	// combination with other strings
	g.Expect(Resolve("{{.Data.key1}}-{{.Data.key2}}", &secret)).To(Equal("value1-value2"))

	// some elaborate pattern
	g.Expect(Resolve("my-{{with $arr := splitN .ObjectMeta.Name \"-\" 3}}{{index $arr 2}}{{end}}-stuff", &secret)).To(Equal("my-secret-foo-bar-stuff"))
}
