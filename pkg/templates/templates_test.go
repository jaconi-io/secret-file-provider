package templates

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRender(t *testing.T) {
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

	// Empty string.
	res, err := Render("", &secret)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(BeEmpty())

	// Invalid pattern.
	res, err = Render("{{ .Data.key1 }", &secret)
	g.Expect(err).To(MatchError("parsing template \"{{ .StringData.key1 }\" failed: template: :1: unexpected \"}\" in operand"))
	g.Expect(res).To(BeEmpty())

	// Missing value in pattern.
	res, err = Render("{{ .Missing }}", &secret)
	g.Expect(err).To(MatchError("executing template \"{{ .Missing }}\" with secret some-namespace/my-super-secret-foo-bar failed: template: :1:3: executing \"\" at <.Missing>: can't evaluate field Missing in type *v1.Secret"))
	g.Expect(res).To(BeEmpty())

	// No pattern.
	res, err = Render("samba-dance", &secret)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(Equal("samba-dance"))

	// Name in pattern.
	res, err = Render("{{ .ObjectMeta.Name }}", &secret)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(Equal("my-super-secret-foo-bar"))

	// Label in pattern.
	res, err = Render("{{ .ObjectMeta.Labels.label1 }}", &secret)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(Equal("labelValue1"))

	// Annotation in pattern.
	res, err = Render("{{ .ObjectMeta.Annotations.annotation1 }}", &secret)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(Equal("annotationValue1"))

	// Secret value in pattern.
	res, err = Render("{{ .Data.key1 }}", &secret)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(Equal("value1"))

	// Combination with other strings.
	res, err = Render("{{ .Data.key1 }}-{{ .Data.key2 }}", &secret)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(Equal("value1-value2"))

	// Some elaborate pattern.
	res, err = Render("my-{{ with $arr := splitN .ObjectMeta.Name \"-\" 3 }}{{ index $arr 2 }}{{ end }}-stuff", &secret)
	g.Expect(err).To(BeNil())
	g.Expect(res).To(Equal("my-secret-foo-bar-stuff"))
}
