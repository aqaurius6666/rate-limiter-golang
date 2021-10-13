package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RemoteURL string

type LimiterServiceServer struct {
	Service   *RedisService
	Logger    *logrus.Logger
	G         *gin.Engine
	RemoteURL RemoteURL
}

func (s *LimiterServiceServer) RegisterProxy() {
	// s.G = gin.Default()
	s.G.Use(gin.Recovery())
	s.G.Use(s.LoggerMiddleware)
	s.G.Use(s.RateLimiterMiddleware)
	s.G.Any("/*proxyPath", s.HandleProxy)
}

//Rate limter middleware
func (s *LimiterServiceServer) RateLimiterMiddleware(c *gin.Context) {
	c.Next()
}

//Logger middleware
func (s *LimiterServiceServer) LoggerMiddleware(c *gin.Context) {
	c.Next()
}

// Proxy hanlder
func (s *LimiterServiceServer) HandleProxy(c *gin.Context) {
	remote, err := url.Parse(string(s.RemoteURL))
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = c.Request.Host
		req.URL.Path = c.Param("proxyPath")
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}
