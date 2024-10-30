package maps_test

import (
	"testing"

	. "github.com/jaconi-io/secret-file-provider/pkg/maps"

	. "github.com/onsi/gomega"
)

func TestToCamel(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(ToCamel("")).To(BeEmpty())
	g.Expect(ToCamel("foo")).To(Equal("Foo"))
	g.Expect(ToCamel("FOO_BAR")).To(Equal("FooBar"))
}

func TestToLowerCamel(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(ToLowerCamel("")).To(BeEmpty())
	g.Expect(ToLowerCamel("foo")).To(Equal("foo"))
	g.Expect(ToLowerCamel("FOO_BAR")).To(Equal("fooBar"))
}
