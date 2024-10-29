package secrets

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestToCamel(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(toCamel("UhH_OHH")).To(Equal("UhhOhh"))
	g.Expect(toCamel("UHH-OHH")).To(Equal("UhhOhh"))
	g.Expect(toCamel("UHH-ohh")).To(Equal("UhhOhh"))
	g.Expect(toCamel("uHH-ohh")).To(Equal("UhhOhh"))
}

func TestToLowerCamel(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(toLowerCamel("UhH_OHH")).To(Equal("uhhOhh"))
	g.Expect(toLowerCamel("UHH-OHH")).To(Equal("uhhOhh"))
	g.Expect(toLowerCamel("UHH-ohh")).To(Equal("uhhOhh"))
	g.Expect(toLowerCamel("uHH-ohh")).To(Equal("uhhOhh"))
}
