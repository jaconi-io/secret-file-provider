package maps

import (
	"strings"

	"github.com/iancoleman/strcase"
)

// ToCamel wraps [strcase.ToCamel], adding support for SCREAMING_SNAKE input.
func ToCamel(val string) string {
	return strcase.ToCamel(strings.ToLower(val))
}

// ToLowerCamel wraps [strcase.ToLowerCamel], adding support for SCREAMING_SNAKE input.
func ToLowerCamel(val string) string {
	return strcase.ToLowerCamel(strings.ToLower(val))
}
