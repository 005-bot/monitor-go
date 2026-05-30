package addressparser

import "strings"

func CleanName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = nonWordRE.ReplaceAllString(name, "")
	name = multiSpaceRE.ReplaceAllString(strings.TrimSpace(name), " ")
	return name
}
