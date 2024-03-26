package web

import (
	"bytes"
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
)

type Server struct {
	config  config.AppConfig
	match   *regexp.Regexp
	replace string
	gicsUrl *url.URL
}

func NewServer(config config.AppConfig) *Server {
	// validate gICS url
	gicsUrl, err := url.Parse(config.Gics.BaseUrl)
	if err != nil {
		slog.Error("Unable to parse 'gics.base-url' config property")
		os.Exit(1)
	}

	return &Server{
		config:  config,
		gicsUrl: gicsUrl,
		match:   regexp.MustCompile(`<consent>`),
		replace: "<notificationClientID>gICS_Soap</notificationClientID><consent>",
	}
}

func (s *Server) Run() error {
	r := s.setupRouter()

	slog.Info("Starting server", "port", s.config.App.Http.Port)
	for _, v := range r.Routes() {
		slog.Info("Route configured", "path", v.Path, "method", v.Method)
	}

	return r.Run(":" + s.config.App.Http.Port)
}

func (s *Server) setupRouter() *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(sloggin.New(slog.Default()), gin.Recovery())

	r.POST("/gics/gicsService", s.handleSoap)
	r.NoRoute(func(c *gin.Context) {
		s.relay(c)
	})

	return r
}

func (s *Server) relay(c *gin.Context) {
	director := func(req *http.Request) {
		req.URL.Scheme = s.gicsUrl.Scheme
		req.URL.Host = s.gicsUrl.Host
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
	slog.Info("Request forwarded", "target", s.gicsUrl.String())
}

func (s *Server) handleSoap(c *gin.Context) {

	body, _ := io.ReadAll(c.Request.Body)
	strBody := string(body)
	newBody := s.match.ReplaceAllString(strBody, s.replace)

	target := s.gicsUrl.String() + "/gics/gicsServiceWithNotification"
	if strBody == newBody {
		s.relay(c)
		return
	}

	// POST to gICS endpoint
	req, err := http.NewRequest("POST", target, bytes.NewBuffer([]byte(newBody)))
	if err != nil {
		slog.Error("Failed to build request", "error", err.Error(), "target", target)
		c.Data(http.StatusBadRequest, "text/plain", []byte(err.Error()))
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to send request", "error", err.Error(), "target", target)
		c.Data(http.StatusBadRequest, "text/plain", []byte(err.Error()))
		return
	}

	gicsResp, _ := io.ReadAll(res.Body)
	c.Data(res.StatusCode, "application/xml", gicsResp)
	slog.Info("Request rewritten", "target", target, "status", res.Status)
}
