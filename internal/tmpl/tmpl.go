package tmpl

import "strings"

// PropToTemplateParam replaces every occurrence of "." with "_" inside prop so that the resulting string could be used as a map key param in a template.
func PropToTemplateParam(prop string) string {
	return strings.ReplaceAll(prop, ".", "_")
}
