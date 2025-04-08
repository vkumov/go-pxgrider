package utils

func IsEmptyJSON(data []byte) bool {
	return string(data) == "{}"
}
