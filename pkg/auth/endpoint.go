package auth

import (
	"regexp"
	"strings"
)

type Endpoint struct {
	scheme     string
	host       string
	owner      string
	repository string
	regexes    []*regexp.Regexp

	TokenHash string
}

func (e *Endpoint) ID() string {
	comps := []string{e.host, e.owner, e.repository}
	return strings.Join(comps, "-")
}
