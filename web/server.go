package web

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/diz-unimr/ths-proxy/notification"
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

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Server struct {
	config  config.AppConfig
	match   *regexp.Regexp
	replace string
	client  HttpClient
	gicsUrl *url.URL
	email   *notification.EmailClient
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
		client:  http.DefaultClient,
		gicsUrl: gicsUrl,
		match:   regexp.MustCompile(`<consent>`),
		replace: "<notificationClientID>gICS_Soap</notificationClientID><consent>",
		email:   notification.NewEmailClient(config.Notification.Email),
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
		c.Data(http.StatusBadRequest, "text/plain", []byte(err.Error()))

		slog.Error("Failed to build request", "error", err.Error(), "target", target)
		return
	}

	res, err := s.client.Do(req)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/plain", []byte(err.Error()))

		slog.Error("Failed to send request", "error", err.Error(), "target", target)
		s.email.Send(err.Error())
		return
	}

	// parse content-type and return data from reader
	ct := res.Header.Get("content-type")
	if ct == "" {
		ct = "text/xml"
	}
	if b, err := io.ReadAll(res.Body); err == nil {
		// return response
		c.Data(res.StatusCode, ct, b)
		slog.Info("Request rewritten", "target", target, "status", res.Status)

		if res.StatusCode >= 400 {
			// send notification
			soapBody := formatResponse(b)
			msg := fmt.Sprintf("gICS responded with error code %d:\n\n%s", res.StatusCode, soapBody)
			s.email.Send(msg)
		}
	}
}

func formatResponse(b []byte) string {
	if formatted, err := xmlIndent(b); err == nil {
		return formatted
	}
	return string(b)
}

func xmlIndent(b []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(b))
	decoder.Strict = false
	buf := new(bytes.Buffer)
	encoder := xml.NewEncoder(buf)
	encoder.Indent("", " ")

tokenize:
	for {
		token, err := decoder.Token()

		switch {
		case err == io.EOF:
			err := encoder.Flush()
			if err != nil {
				return "", err
			}

			break tokenize
		case err != nil:
			slog.Debug("Failed to tokenize xml", "error", err)
			return "", err
		default:
			err = encoder.EncodeToken(token)
			if err != nil {
				slog.Debug("Failed to encode xml", "error", err)
				return "", err
			}

		}
	}

	return buf.String(), nil
}
