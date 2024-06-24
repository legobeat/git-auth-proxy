package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/go-crypt/crypt"
	"github.com/legobeat/git-auth-proxy/pkg/config"
)

type Provider interface {
	getPathRegex(owner, repository string) ([]*regexp.Regexp, error)
	getAuthorizationHeader(ctx context.Context, path string) (string, error)
	getHost(e *Endpoint, path string) string
	getPath(e *Endpoint, path string) string
}

type Authorizer struct {
	providers        map[string]Provider
	endpoints        []*Endpoint
	endpointsByID    map[string]*Endpoint
	endpointsByToken map[string]*Endpoint
}

func NewAuthorizer(cfg *config.Configuration) (*Authorizer, error) {
	providers := map[string]Provider{}
	endpoints := []*Endpoint{}
	endpointsByID := map[string]*Endpoint{}
	endpointsByToken := map[string]*Endpoint{}

	for _, p := range cfg.Policies {
		// Get the correct provider for the policy
		var provider Provider
		switch p.Provider {
		case config.GitHubProviderType:
			ghToken := p.GitHub.Token
			provider = newGithub(ghToken)
		default:
			return nil, fmt.Errorf("invalid provider type %s", p.Provider)
		}

		regexes := make([]*regexp.Regexp, 0)

		// Create endpoint for the repositories
		for _, r := range p.Repositories {
			pathRegex, err := provider.getPathRegex(r.Owner, r.Name)
			if err != nil {
				return nil, fmt.Errorf("could not get path regex: %w", err)
			}

			regexes = append(regexes, pathRegex...)
		}
		e := &Endpoint{
			host:      p.Host,
			scheme:    p.Scheme,
			id:        p.ID,
			regexes:   regexes,
			TokenHash: p.UserAuth.TokenHash,
		}
		providers[e.ID()] = provider
		endpoints = append(endpoints, e)
		endpointsByID[e.ID()] = e
		endpointsByToken[p.UserAuth.TokenHash] = e
	}

	authz := &Authorizer{
		providers:        providers,
		endpoints:        endpoints,
		endpointsByID:    endpointsByID,
		endpointsByToken: endpointsByToken,
	}
	return authz, nil
}

func (a *Authorizer) GetEndpoints() []*Endpoint {
	return a.endpoints
}

func (a *Authorizer) GetEndpointById(id string) (*Endpoint, error) {
	e, ok := a.endpointsByID[id]
	if !ok {
		return nil, fmt.Errorf("endpoint not found for id %s", id)
	}
	return e, nil
}

func (a *Authorizer) GetEndpointByToken(token string) (*Endpoint, error) {
	for tokenHash, e := range a.endpointsByToken {
		// empty hash = anon policy. skip CheckPassword.
		if tokenHash == "" && token == "" {
			return e, nil
		}
		valid, err := crypt.CheckPassword(token, tokenHash)
		if err != nil {
			panic(err)
		}
		if valid {
			return e, nil
		}
	}
	if token == "" {
		return nil, fmt.Errorf("missing basic auth")
	}
	return nil, fmt.Errorf("endpoint not found for given token")
}

func (a *Authorizer) IsPermitted(path string, token string) error {
	e, err := a.GetEndpointByToken(token)
	if err != nil {
		return err
	}
	for _, r := range e.regexes {
		if r.MatchString(path) {
			return nil
		}
	}
	return fmt.Errorf("token not permitted for path %s", path)
}

func (a *Authorizer) UpdateRequest(ctx context.Context, req *http.Request, token string) (*http.Request, *url.URL, error) {
	e, err := a.GetEndpointByToken(token)
	if err != nil {
		return nil, nil, err
	}
	provider, ok := a.providers[e.ID()]
	if !ok {
		return nil, nil, fmt.Errorf("provider not found for id %s", e.ID())
	}

	host := provider.getHost(e, req.URL.Path)
	path := provider.getPath(e, req.URL.Path)
	authorizationHeader, err := provider.getAuthorizationHeader(ctx, req.URL.Path)
	if err != nil {
		return nil, nil, err
	}
	url, err := url.Parse(fmt.Sprintf("%s://%s", e.scheme, host))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid url format: %w", err)
	}

	req.Host = host
	req.URL.Path = path
	req.Header.Del("Authorization")
	req.Header.Add("Authorization", authorizationHeader)
	return req, url, nil
}
