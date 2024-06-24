package auth

import (
	"regexp"
	"strings"
)

type Endpoint struct {
	scheme       string
	host         string
	organization string
	repository   string
	regexes      []*regexp.Regexp

	Token      string
	SecretName string
}

func (e *Endpoint) ID() string {
	comps := []string{e.host, e.organization}
	comps = append(comps, e.repository)
	return strings.Join(comps, "-")
}
