package auth

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

const standardGitHub = "github.com"

type GitHubTokenSource interface {
	Token(ctx context.Context) (string, error)
}

type github struct {
	itr GitHubTokenSource
}

type githubDummyTokenSource struct {
	token string
}

func (s githubDummyTokenSource) Token(_ctx context.Context) (string, error) {
	return s.token, nil
}

func newGithub(token string) *github {
	itr := githubDummyTokenSource{
		token: token,
	}
	return &github{itr: itr}
}

func (g *github) getPathRegex(owner, repository string) ([]*regexp.Regexp, error) {
	// Support wildcards
	if owner == "*" || owner == "" {
		owner = "[^/]*"
	}
	// Support wildcards
	if repository == "*" || repository == "" {
		repository = "[^/]*"
	}
	git, err := regexp.Compile(fmt.Sprintf(`(?i)/%s/%s(/.*)?\b`, owner, repository))
	if err != nil {
		return nil, err
	}
	api, err := regexp.Compile(fmt.Sprintf(`(?i)/api/v[23]/(.*)/%s/%s/(/.*)?\b`, owner, repository))
	if err != nil {
		return nil, err
	}
	repos, err := regexp.Compile(fmt.Sprintf(`(?i)/repos/(.*)/%s/%s/(/.*)?\b`, owner, repository))
	if err != nil {
		return nil, err
	}
	// static graphql regexp will be duplicated across policies despite being redundant - room for improvement
	graphql := regexp.MustCompile(`(?i)/graphql/?\b`)
	if err != nil {
		return nil, err
	}
	return []*regexp.Regexp{git, api, repos, graphql}, nil
}

func (g *github) getAuthorizationHeader(ctx context.Context, path string) (string, error) {
	token, err := g.itr.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("error when fetching GitHub token: %w", err)
	}
	if token == "" {
		return "", nil
	}

	if strings.HasPrefix(path, "/api/v3/") {
		return fmt.Sprintf("Bearer %s", token), nil
	}
	if strings.HasPrefix(path, "/graphql") {
		return fmt.Sprintf("bearer %s", token), nil
	}
	tokenB64 := b64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("x-access-token:%s", token)))
	return fmt.Sprintf("Basic %s", tokenB64), nil
}

func (g *github) getHost(e *Endpoint, path string) string {
	if e.host != standardGitHub {
		return e.host
	}
	if strings.HasPrefix(path, "/api/v3/") || strings.HasPrefix(path, "/repos/") || strings.HasPrefix(path, "/graphql") {
		return fmt.Sprintf("api.%s", e.host)
	}
	return e.host
}

func (g *github) getPath(e *Endpoint, path string) string {
	if e.host != standardGitHub {
		return path
	}
	newPath := strings.TrimPrefix(path, "/api/v3")
	return newPath
}
