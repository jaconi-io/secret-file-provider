package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testString = `foo:
  bar:
    baz: 42
  oof: 7
`

var testData = map[interface{}]interface{}{
	"foo": map[interface{}]interface{}{
		"bar": map[interface{}]interface{}{
			"baz": 42,
		},
		"oof": 7,
	},
}

type invalidYAML struct{}

func (*invalidYAML) MarshalYAML() (interface{}, error) {
	return nil, errors.New("expected")
}

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

func TestReadAllMissing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	defer viper.Reset()
	viper.Set(env.SecretFileSingle, false)

	content, err := ReadAll("/foo/bar")
	g.Expect(err).To(gomega.MatchError(os.IsNotExist, "IsNotExist"))
	g.Expect(content).To(gomega.BeNil())
}

func TestReadAllInvalidContent(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, false)
	defer viper.Reset()

	f, err := os.CreateTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	err = os.WriteFile(f.Name(), []byte(`invalid`), 0644)
	g.Expect(err).To(gomega.BeNil())

	content, err := ReadAll(f.Name())
	g.Expect(err).To(gomega.MatchError("yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid` into map[interface {}]interface {}"))
	g.Expect(content).To(gomega.BeNil())
}

func TestReadAll(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, false)
	defer viper.Reset()

	f, err := os.CreateTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	err = os.WriteFile(f.Name(), []byte(testString), 0644)
	g.Expect(err).To(gomega.BeNil())

	content, err := ReadAll(f.Name())
	g.Expect(err).To(gomega.BeNil())
	g.Expect(content["foo"]).To(gomega.Equal(testData["foo"]))
}

func TestReadAllFilePerSecretMissingDirectory(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, true)
	defer viper.Reset()

	content, err := ReadAll("/foo/bar")
	g.Expect(err).To(gomega.MatchError(os.IsNotExist, "IsNotExist"))
	g.Expect(content).To(gomega.BeNil())
}

func TestReadAllFilePerSecretEmptyDirectory(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, true)
	defer viper.Reset()

	dir, err := os.MkdirTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	content, err := ReadAll(dir)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(content).To(gomega.BeEmpty())
}

func TestReadAllFilePerSecretUnreadableFile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, true)
	defer viper.Reset()

	dir, err := os.MkdirTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	f, err := os.CreateTemp(dir, "foo")
	g.Expect(err).To(gomega.BeNil())

	err = f.Chmod(0000)
	g.Expect(err).To(gomega.BeNil())

	content, err := ReadAll(dir)
	g.Expect(err).To(gomega.MatchError(os.IsPermission, "IsPermission"))
	g.Expect(content).To(gomega.BeEmpty())
}

func TestReadAllFilePerSecret(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, true)
	defer viper.Reset()

	dir, err := os.MkdirTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	f1, err := os.CreateTemp(dir, "foo")
	g.Expect(err).To(gomega.BeNil())

	f2, err := os.CreateTemp(dir, "bar")
	g.Expect(err).To(gomega.BeNil())

	err = os.WriteFile(f1.Name(), []byte(testString), 0644)
	g.Expect(err).To(gomega.BeNil())

	err = os.WriteFile(f2.Name(), []byte(testString), 0644)
	g.Expect(err).To(gomega.BeNil())

	content, err := ReadAll(dir)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(content).To(gomega.Equal(map[interface{}]interface{}{
		filepath.Base(f1.Name()): testString,
		filepath.Base(f2.Name()): testString,
	}))
}

func TestWriteAllMkdirForbidden(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, false)
	defer viper.Reset()

	parent, err := os.MkdirTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	err = os.Chmod(parent, 0000)
	g.Expect(err).To(gomega.BeNil())

	err = WriteAll(path.Join(parent, "bar", "baz"), testData)
	g.Expect(err).To(gomega.MatchError(os.IsPermission, "IsPermission"))
}

func TestWriteAllFileForbidden(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, false)
	defer viper.Reset()

	f, err := os.CreateTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	err = f.Chmod(0000)
	g.Expect(err).To(gomega.BeNil())

	err = WriteAll(f.Name(), testData)
	g.Expect(err).To(gomega.MatchError(os.IsPermission, "IsPermission"))
}

func TestWriteAllInvalid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, false)
	defer viper.Reset()

	f, err := os.CreateTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	err = WriteAll(f.Name(), map[interface{}]interface{}{
		"invalid": &invalidYAML{},
	})
	g.Expect(err).To(gomega.MatchError(fmt.Sprintf("invalid secret content for %s: expected", f.Name())))
}

func TestWriteAll(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, false)
	defer viper.Reset()

	f, err := os.CreateTemp("", "bar")
	g.Expect(err).To(gomega.BeNil())

	err = WriteAll(f.Name(), testData)
	g.Expect(err).To(gomega.BeNil())

	b, err := io.ReadAll(f)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(string(b)).To(gomega.Equal(testString))
}

func TestWriteAllFilePerSecretMkdirForbidden(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, true)
	defer viper.Reset()

	parent, err := os.MkdirTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	err = os.Chmod(parent, 0000)
	g.Expect(err).To(gomega.BeNil())

	err = WriteAll(path.Join(parent, "bar"), testData)
	g.Expect(err).To(gomega.MatchError(os.IsPermission, "IsPermission"))
}

func TestWriteAllFilePerSecretFileForbidden(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, true)
	defer viper.Reset()

	dir, err := os.MkdirTemp("", "foo")
	g.Expect(err).To(gomega.BeNil())

	f, err := os.CreateTemp(dir, "foo")
	g.Expect(err).To(gomega.BeNil())

	err = f.Chmod(0000)
	g.Expect(err).To(gomega.BeNil())

	err = WriteAll(dir, map[interface{}]interface{}{
		filepath.Base(f.Name()): testData,
	})
	g.Expect(err).To(gomega.MatchError(os.IsPermission, "IsPermission"))
}

func TestWriteAllFilePerSecret(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	viper.Set(env.SecretFileSingle, true)
	defer viper.Reset()

	dir, err := os.MkdirTemp("", "bar")
	g.Expect(err).To(gomega.BeNil())

	err = WriteAll(dir, map[interface{}]interface{}{
		"foo": testString,
		"bar": testString,
	})
	g.Expect(err).To(gomega.BeNil())

	b, err := os.ReadFile(filepath.Join(dir, "foo"))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(string(b)).To(gomega.Equal(testString))

	b, err = os.ReadFile(filepath.Join(dir, "bar"))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(string(b)).To(gomega.Equal(testString))
}
