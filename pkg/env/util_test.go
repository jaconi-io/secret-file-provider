package env_test

import (
	"testing"

	. "github.com/jaconi-io/secret-file-provider/pkg/env"

	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func TestGetFinalizer(t *testing.T) {
	g := NewGomegaWithT(t)

	defer viper.Reset()
	viper.Set(PodName, "some-unusually-lengthy-pod-name-6d98ccb7dd-c8zr8")

	g.Expect(GetFinalizer()).To(HaveLen(63))
	g.Expect(GetFinalizer()).To(Equal("jaconi.io/secret-file-provider-engthy-pod-name-6d98ccb7dd-c8zr8"))
}
