// As strcase.ToCamel does not work for SCREAMING_SNAKE input, we need our own function here
package maps

import (
	"strings"

	"github.com/iancoleman/strcase"
)

func ToCamel(val string) string {
	return strcase.ToCamel(strings.ToLower(val))
}

func ToLowerCamel(val string) string {
	return strcase.ToLowerCamel(strings.ToLower(val))
}
