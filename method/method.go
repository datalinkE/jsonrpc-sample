package method

import (
	"strings"
)

func PathHasMethod(path string, method string) bool {
	pathWords := strings.Split(path, "/")
	if method == "" || len(pathWords) < 1 {
		return false
	}

	return pathWords[len(pathWords)-1] == method
}
