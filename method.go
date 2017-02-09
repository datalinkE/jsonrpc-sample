package rpcserver

import (
	"strings"
)

func PathHasMethod(path string, method string) bool {
	pathLastPart := LastPart(path)
	return pathLastPart == method
}

func LastPart(path string) string {
	pathWords := strings.Split(path, "/")
	length := len(pathWords)
	if length == 0 { // should not be
		return path
	}
	return pathWords[length-1]
}
