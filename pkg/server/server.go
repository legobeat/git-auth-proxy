package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	pkggin "github.com/xenitab/pkg/gin"

	"github.com/legobeat/git-auth-proxy/pkg/auth"
)

type GitProxy struct {
	authz *auth.Authorizer
}

func NewGitProxy(authz *auth.Authorizer) *GitProxy {
	return &GitProxy{
		authz: authz,
	}
}

func (g *GitProxy) Server(ctx context.Context, addr string) *http.Server {
	cfg := pkggin.DefaultConfig()
	cfg.LogConfig.Logger = logr.FromContextOrDiscard(ctx)
	cfg.MetricsConfig.HandlerID = "proxy"
	router := pkggin.NewEngine(cfg)
	router.GET("/readyz", readinessHandler)
	router.GET("/healthz", livenessHandler)
	router.NoRoute(g.proxyHandler)
	// The ReadTimeout is set to 5 min make sure that strange requests don't live forever
	// But in general the external request should set a good timeout value for it's request.
	srv := &http.Server{ReadTimeout: 5 * time.Minute, Addr: addr, Handler: router}
	return srv
}

func (g *GitProxy) proxyHandler(c *gin.Context) {
	// Get the token from the request
	// error is fine; we fall back to "", the default public policy, if any
	//nolint: ineffassign //ignore
	token, err := getTokenFromRequest(c.Request)
	// Check basic auth with local auth configuration
	err = g.authz.IsPermitted(c.Request.URL.EscapedPath(), token)
	if err != nil {
		//nolint: errcheck //ignore
		c.Error(fmt.Errorf("Received unauthorized request: %w", err))
		c.String(http.StatusForbidden, "User not permitted")
		return
	}
	// Authenticate the request with the proper token
	req, url, err := g.authz.UpdateRequest(c.Request.Context(), c.Request, token)
	if err != nil {
		//nolint: errcheck //ignore
		c.Error(fmt.Errorf("Could not authenticate request: %w", err))
		c.String(http.StatusInternalServerError, "Internal server error")
		return
	}

	// TODO (Philip): Add caching of the proxy
	// Forward the request to the correct proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(c.Writer, req)
}

func readinessHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

func livenessHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}
