package fetchsms

import (
	"errors"
	"net/url"
)

type Tab string

const (
	TabActive  Tab = "active"
	TabHistory Tab = "history"
)

func tabQuery(tab Tab) (url.Values, error) {
	if tab == "" {
		tab = TabActive
	}
	if tab != TabActive && tab != TabHistory {
		return nil, errors.New("fetchsms: invalid tab")
	}
	return url.Values{"tab": {string(tab)}}, nil
}
