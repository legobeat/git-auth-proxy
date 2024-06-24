package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/legobeat/git-auth-proxy/pkg/config"
	"github.com/stretchr/testify/require"
)

type MockGitHubTokenSource struct {
}

func (*MockGitHubTokenSource) Token(ctx context.Context) (string, error) {
	return "foo", nil
}

func getGitHubAuthorizerSingle() *Authorizer {
	cfg := &config.Configuration{
		Policies: []*config.Policy{
			{
				ID:       "123",
				Provider: config.GitHubProviderType,
				GitHub: config.GitHub{
					Token: "test-token",
				},
				Host: "github.com",
				Repositories: []*config.Repository{
					{
						Owner: "org",
						Name:  "repo",
					},
					{
						Owner: "org",
						Name:  "foobar",
					},
					{
						Owner: "org",
						Name:  "repo%20space",
					},
				},
				UserAuth: config.UserAuth{
					TokenHash: "$6$NmUowWy4LgRFWSsY$fOVzziH1IYD84dW8qSHa4X9PSHlo4R52oTx4jzvrR5vWkepDM/sWC.zbgrZ1IZ90zBoUGoEGCLQdbpaMbWtou.",
					// mkpasswd -m sha512crypt incoming-test-token
				},
			},
		},
	}
	auth, err := NewAuthorizer(cfg)
	if err != nil {
		panic(err)
	}
	return auth
}

func getGitHubAuthorizerMixed() *Authorizer {
	cfg := &config.Configuration{
		Policies: []*config.Policy{
			{
				ID:       "private",
				Provider: config.GitHubProviderType,
				GitHub: config.GitHub{
					Token: "test-token",
				},
				Host: "github.com",
				Repositories: []*config.Repository{
					{
						Owner: "org",
						Name:  "repo",
					},
					{
						Owner: "org",
						Name:  "foobar",
					},
					{
						Owner: "org",
						Name:  "repo%20space",
					},
				},
				UserAuth: config.UserAuth{
					TokenHash: "$6$SuRwqTHhR/k4axcK$4THoTcwS75DCmwQ5YRC8sWoD0g/pXSHxa0fam00TsOE.6UNMUa0N9.XTcWtH6171MM.BRV7QMomEwBdh57GfN1",
				},
			},
			{
				ID:       "public",
				Provider: config.GitHubProviderType,
				GitHub: config.GitHub{
					Token: "",
				},
				Host: "github.com",
				Repositories: []*config.Repository{
					{
						Owner: "*",
						Name:  "*",
					},
				},
				UserAuth: config.UserAuth{
					TokenHash: "",
				},
			},
		},
	}
	auth, err := NewAuthorizer(cfg)
	if err != nil {
		panic(err)
	}
	return auth
}

func TestGitHubMixedAuthorization(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		allow bool
	}{
		{
			name:  "allow root",
			path:  "/",
			allow: true,
		},
		{
			name:  "allow repo",
			path:  "/org/repo",
			allow: true,
		},
		{
			name:  "allow repo",
			path:  "/Org/repO",
			allow: true,
		},
		{
			name:  "allow api",
			path:  "/api/v3/org/repo",
			allow: true,
		},
		{
			name:  "allow graphql",
			path:  "/graphql",
			allow: true,
		},
		{
			name:  "allow catchall repo",
			path:  "/org/foo",
			allow: true,
		},
		{
			name:  "allow catchall repo in api",
			path:  "/api/v3/org/foo",
			allow: true,
		},
		{
			name:  "allow catchall org",
			path:  "/foo/repo",
			allow: true,
		},
		{
			name:  "allow catchall org in api",
			path:  "/api/v3/foo/repo",
			allow: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := getGitHubAuthorizerMixed()
			endpoint, err := authz.GetEndpointById("github.com//private")
			require.NotNil(t, endpoint)
			require.NoError(t, err)
			err = authz.IsPermitted(tt.path, "private-test-token")

			if tt.allow {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGitHubAuthorization(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		allow bool
	}{
		{
			name:  "allow root",
			path:  "/",
			allow: true,
		},
		{
			name:  "allow repo",
			path:  "/org/repo",
			allow: true,
		},
		{
			name:  "allow repo",
			path:  "/Org/repO",
			allow: true,
		},
		{
			name:  "allow api",
			path:  "/api/v3/org/repo",
			allow: true,
		},
		{
			name:  "allow graphql",
			path:  "/graphql",
			allow: true,
		},
		{
			name:  "disallow wrong repo",
			path:  "/org/foo",
			allow: false,
		},
		{
			name:  "disallow wrong repo in api",
			path:  "/api/v3/org/foo",
			allow: false,
		},
		{
			name:  "disallow wrong org",
			path:  "/foo/repo",
			allow: false,
		},
		{
			name:  "disallow wrong org in api",
			path:  "/api/v3/foo/repo",
			allow: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := getGitHubAuthorizerSingle()
			endpoint, err := authz.GetEndpointById("github.com//123")
			require.NotNil(t, endpoint)
			require.NoError(t, err)
			err = authz.IsPermitted(tt.path, "incoming-test-token")

			if tt.allow {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGithubApiGetAuthorization(t *testing.T) {
	gh := &github{itr: &MockGitHubTokenSource{}}
	authorization, err := gh.getAuthorizationHeader(context.TODO(), "/api/v3/test")
	require.NoError(t, err)
	require.Equal(t, "Bearer foo", authorization)
}

func TestGithubGraphqlGetAuthorization(t *testing.T) {
	gh := &github{itr: &MockGitHubTokenSource{}}
	authorization, err := gh.getAuthorizationHeader(context.TODO(), "/graphql")
	require.NoError(t, err)
	require.Equal(t, "bearer foo", authorization)
}

func TestGithubGitGetAuthorization(t *testing.T) {
	gh := &github{itr: &MockGitHubTokenSource{}}
	authorization, err := gh.getAuthorizationHeader(context.TODO(), "/org/repo")
	require.NoError(t, err)
	require.Equal(t, "Basic eC1hY2Nlc3MtdG9rZW46Zm9v", authorization)
}

func TestGetHost(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		path     string
		expected string
	}{
		{
			name:     "api path standard github",
			host:     standardGitHub,
			path:     "/api/v3/test",
			expected: fmt.Sprintf("api.%s", standardGitHub),
		},
		{
			name:     "repo path standard github",
			host:     standardGitHub,
			path:     "/foo/bar",
			expected: standardGitHub,
		},
		{
			name:     "api path enterprise github",
			host:     "example.com",
			path:     "/api/v3/test",
			expected: "example.com",
		},
		{
			name:     "repo path enterprise github",
			host:     "example.com",
			path:     "/foo/bar",
			expected: "example.com",
		},
	}
	gh := &github{itr: &MockGitHubTokenSource{}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoint{
				host: tt.host,
			}
			host := gh.getHost(e, tt.path)
			require.Equal(t, tt.expected, host)
		})
	}
}

func TestGetPath(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		path     string
		expected string
	}{
		{
			name:     "api path standard github",
			host:     standardGitHub,
			path:     "/api/v3/test",
			expected: "/test",
		},
		{
			name:     "repo path standard github",
			host:     standardGitHub,
			path:     "/foo/bar",
			expected: "/foo/bar",
		},
		{
			name:     "api path enterprise github",
			host:     "example.com",
			path:     "/api/v3/test",
			expected: "/api/v3/test",
		},
		{
			name:     "repo path enterprise github",
			host:     "example.com",
			path:     "/foo/bar",
			expected: "/foo/bar",
		},
	}
	gh := &github{itr: &MockGitHubTokenSource{}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoint{
				host: tt.host,
			}
			path := gh.getPath(e, tt.path)
			require.Equal(t, tt.expected, path)
		})
	}
}

func TestGetEndpointByToken(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		expectedId string
	}{
		{
			name:       "valid token: private endpoints",
			token:      "private-test-token",
			expectedId: "github.com//private",
		},
		{
			name:       "invalid token: public endpoints",
			token:      "invalid-token",
			expectedId: "github.com//public",
		},
		{
			name:       "no token: public endpoints",
			token:      "",
			expectedId: "github.com//public",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := getGitHubAuthorizerMixed()
			endpoint, err := authz.GetEndpointByToken(tt.token)
			require.NotNil(t, endpoint)
			require.NoError(t, err)
			require.Equal(t, tt.expectedId, endpoint.ID())
		})
	}
}
