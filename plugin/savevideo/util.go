package savevideo

import "strings"

func IsErrorAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "already exists")
}
