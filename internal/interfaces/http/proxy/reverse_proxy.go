package proxy

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func NewReverseProxy(target string, basePath string) gin.HandlerFunc {
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(c *gin.Context) {
		// Exemple: /api/v1/plans/123 -> /plans/123
		path := c.Request.URL.Path
		c.Request.URL.Path = path[len(basePath):]

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
