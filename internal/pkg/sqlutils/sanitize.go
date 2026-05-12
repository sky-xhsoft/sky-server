package sqlutils

import (
	"fmt"
	"regexp"
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func SanitizeIdentifier(name string) (string, error) {
	if !validIdentifier.MatchString(name) {
		return "", fmt.Errorf("invalid SQL identifier: %s", name)
	}
	return name, nil
}
