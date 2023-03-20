package secrets

import (
	"testing"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReadSecretContent_wholeContent(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	defer viper.Reset()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "1-2-3-4",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Data: map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		},
	}

	// attach on root level
	result := readSecretContent(secret)
	g.Expect(result).To(gomega.Equal(map[interface{}]interface{}{"key1": "value1", "key2": "value2"}))

	// with simple property path
	viper.Set(env.SecretFilePropertyPattern, "foo")
	result = readSecretContent(secret)
	g.Expect(result).To(gomega.Equal(map[interface{}]interface{}{"foo": map[interface{}]interface{}{"key1": "value1", "key2": "value2"}}))

	// with templated property path
	viper.Set(env.SecretFilePropertyPattern, "{{.ObjectMeta.Labels.foo}}")
	result = readSecretContent(secret)
	g.Expect(result).To(gomega.Equal(map[interface{}]interface{}{"bar": map[interface{}]interface{}{"key1": "value1", "key2": "value2"}}))
}

func TestReadSecretContent_singleSelect(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	defer viper.Reset()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "1-2-3-4",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Data: map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		},
	}
	viper.Set(env.SecretContentSelector, "{{.Data.key1}}")

	// attach on root level
	result := readSecretContent(secret)
	g.Expect(result).To(gomega.Equal(map[interface{}]interface{}{"key1": "value1"}))

	// with simple property path
	viper.Set(env.SecretFilePropertyPattern, "foo")
	result = readSecretContent(secret)
	g.Expect(result).To(gomega.Equal(map[interface{}]interface{}{"foo": "value1"}))

	// with templated property path
	viper.Set(env.SecretFilePropertyPattern, "{{.ObjectMeta.Labels.foo}}")
	result = readSecretContent(secret)
	g.Expect(result).To(gomega.Equal(map[interface{}]interface{}{"bar": "value1"}))
}
