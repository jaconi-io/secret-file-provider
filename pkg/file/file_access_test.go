package file

import (
	"os"
	"testing"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/logger"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testString = `
foo:
  bar:
    baz: 42
  oof: 7
`

func TestName(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	defer viper.Reset()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo-bar",
		},
	}

	// plain string
	viper.Set(env.SecretFileNamePattern, "/var/config/secret-samba.yaml")
	g.Expect(Name(secret)).To(gomega.Equal("/var/config/secret-samba.yaml"))

	// simple template
	viper.Set(env.SecretFileNamePattern, "/var/config/secret-{{.ObjectMeta.Name}}.yaml")
	g.Expect(Name(secret)).To(gomega.Equal("/var/config/secret-foo-bar.yaml"))

	// a more elaborate template...
	secret.ObjectMeta.Name = "1-2-3-4-5"
	viper.Set(env.SecretFileNamePattern, "/var/config/secret-{{with $arr := splitN .ObjectMeta.Name \"-\" 3}}{{index $arr 2}}{{end}}.yaml")
	g.Expect(Name(secret)).To(gomega.Equal("/var/config/secret-3-4-5.yaml"))
}

func TestReadAll(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	log := logger.New(&corev1.Secret{})

	content := ReadAll(log, "/dowhap/dododowhap")
	g.Expect(content).To(gomega.BeEmpty())

	f, _ := os.CreateTemp("", "foo")
	os.WriteFile(f.Name(), []byte(testString), 0644)

	content = ReadAll(log, f.Name())
	g.Expect(content["foo"]).To(gomega.Equal(map[interface{}]interface{}{"bar": map[interface{}]interface{}{"baz": 42}, "oof": 7}))
}
