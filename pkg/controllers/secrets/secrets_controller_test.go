package secrets

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	testfile string
	req      = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
)

func TestReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	defer viper.Reset()
	defer os.Remove(testfile)
	viper.Set(env.SecretContentSelector, "{{.Data.key1}}")
	viper.Set(env.SecretFileNamePattern, testfile)
	viper.Set(env.SecretFilePropertyPattern, "{{.ObjectMeta.Labels.company}}")

	// Create file and add content
	secret1 := testSecret("acme")
	reconciler := &Reconciler{Client: fake.NewFakeClient(secret1)}

	_, err := reconciler.Reconcile(context.TODO(), req)
	g.Expect(err).To(BeNil())

	// verify file existing and has proper content
	result := readTestFile()
	g.Expect(err).To(BeNil())
	g.Expect(result).To(HaveLen(1))
	g.Expect(result["acme"]).To(Equal("value1"))

	// Append content to file
	viper.Set(env.SecretContentSelector, "{{.Data.key2}}")
	secret2 := testSecret("company")
	reconciler = &Reconciler{Client: fake.NewFakeClient(secret2)}

	_, err = reconciler.Reconcile(context.TODO(), req)
	g.Expect(err).To(BeNil())

	// verify that new property was added
	result = readTestFile()

	g.Expect(result).To(HaveLen(2))
	g.Expect(result["acme"]).To(Equal("value1"))
	g.Expect(result["company"]).To(Equal("value2"))

	// Remove property from file
	secret1.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	reconciler = &Reconciler{Client: fake.NewFakeClient(secret1)}
	_, err = reconciler.Reconcile(context.TODO(), req)
	g.Expect(err).To(BeNil())

	// verify that property has gone
	result = readTestFile()

	g.Expect(result).To(HaveLen(1))
	g.Expect(result["acme"]).To(BeNil())
	g.Expect(result["company"]).To(Equal("value2"))

	// Change content selector to push all
	viper.Set(env.SecretContentSelector, "")
	secret3 := testSecret("uni")
	reconciler = &Reconciler{Client: fake.NewFakeClient(secret3)}

	_, err = reconciler.Reconcile(context.TODO(), req)
	g.Expect(err).To(BeNil())

	// verify that new property tree was added
	result = readTestFile()

	g.Expect(result["uni"]).To(Equal(map[interface{}]interface{}{"key1": "value1", "key2": "value2"}))
}

func readTestFile() map[interface{}]interface{} {
	bytes, err := os.ReadFile(testfile)
	if err != nil {
		panic(fmt.Sprintf("Failed to read file %s", testfile))
	}

	result := make(map[interface{}]interface{})
	err = yaml.Unmarshal(bytes, result)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal content for file %s", testfile))
	}
	return result
}

func testSecret(company string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels: map[string]string{
				"company": company,
			},
		},
		Data: map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		},
	}
}

func randStringBytes(n int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func init() {
	testfile = filepath.Join(os.TempDir(), randStringBytes(10)+".yaml")
}
