package fetchsms

import (
	"errors"
	"regexp"
)

var resourcePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func resourcePath(collection, id, suffix string) (string, error) {
	if !resourcePattern.MatchString(id) {
		return "", errors.New("fetchsms: invalid resource ID")
	}
	return "/" + collection + "/" + id + suffix, nil
}
