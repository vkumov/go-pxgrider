package mappings

import "strings"

func cleanJSONSchema(schema string) string {
	return strings.Replace(strings.Replace(schema, "\t", "", -1), "\n", "", -1)
}
