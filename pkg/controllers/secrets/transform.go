package secrets

import (
	"strings"

	"github.com/iancoleman/strcase"
)

var keyTransformFunctions = map[string]func(string) string{
	"ToCamel":          toCamel,
	"ToLowerCamel":     toLowerCamel,
	"ToKebab":          strcase.ToKebab,
	"ToScreamingKebab": strcase.ToScreamingKebab,
	"ToSnake":          strcase.ToSnake,
	"ToScreamingSnake": strcase.ToScreamingSnake,
}

func toCamel(val string) string {
	return strcase.ToCamel(strings.ToLower(val))
}

func toLowerCamel(val string) string {
	return strcase.ToLowerCamel(strings.ToLower(val))
}
