package auth

import (
	"regexp"
	"strings"
)

type Endpoint struct {
	scheme     string
	host       string
	owner      string
	project    string
	repository string
	regexes    []*regexp.Regexp

	TokenHash  string
	Namespaces []string
}

func (e *Endpoint) ID() string {
	comps := []string{e.host, e.owner}
	if e.project != "" {
		comps = append(comps, e.project)
	}
	comps = append(comps, e.repository)
	return strings.Join(comps, "-")
}
