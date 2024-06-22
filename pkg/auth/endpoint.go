package auth

import (
	"regexp"
	"strings"
)

type Endpoint struct {
	scheme  string
	host    string
	id      string
	regexes []*regexp.Regexp

	TokenHash string
}

func (e *Endpoint) ID() string {
	comps := []string{e.host, e.id}
	return strings.Join(comps, "-")
}
