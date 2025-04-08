package mappings

import (
	"fmt"
	"strings"
)

func (m ServiceRestMappings) GetMapping(method string) (RestMapping, string, error) {
	if mapping, ok := m[method]; ok {
		return mapping, "", nil
	}

	for k, v := range m {
		if strings.ToLower(k) == strings.ToLower(method) {
			return v, k, nil
		}
	}

	return RestMapping{}, "", fmt.Errorf("mo mapping found for method %s", method)
}
