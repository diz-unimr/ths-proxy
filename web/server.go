package web

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/diz-unimr/ths-proxy/notification"
	"github.com/gin-gonic/gin"
	"github.com/go-xmlfmt/xmlfmt"
	sloggin "github.com/samber/slog-gin"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
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
	proxy   *httputil.ReverseProxy
}

func NewServer(config config.AppConfig) *Server {
	// validate gICS url
	gicsUrl, err := url.Parse(config.Gics.BaseUrl)
	if err != nil {
		slog.Error("Unable to parse 'gics.base-url' config property")
		os.Exit(1)
	}

	proxy := httputil.NewSingleHostReverseProxy(gicsUrl)
	proxy.Transport = &NotifyTransport{
		notifier: notification.NewEmailClient(config.Notification.Email),
		match:    config.Notification.MatchService,
	}

	return &Server{
		config:  config,
		client:  http.DefaultClient,
		gicsUrl: gicsUrl,
		proxy:   proxy,
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
	r.NoRoute(s.relay)
	return r
}

func (s *Server) relay(c *gin.Context) {

	s.proxy.ServeHTTP(c.Writer, c.Request)
	slog.Info("Request forwarded", "target", s.gicsUrl.String())
}

func (s *Server) handleSoap(c *gin.Context) {

	body, _ := io.ReadAll(c.Request.Body)
	strBody := string(body)
	newBody := s.match.ReplaceAllString(strBody, s.replace)
	if strBody == newBody {
		s.relay(c)
	}

	target := s.gicsUrl.String() + "/gics/gicsServiceWithNotification"
	targetUrl, err := url.Parse(target)
	if err != nil {
		s.relay(c)
		return
	}
	c.Request.URL = targetUrl

	// rewrite body
	c.Request.Body = io.NopCloser(strings.NewReader(newBody))
	c.Request.ContentLength = int64(len(newBody))

	s.relay(c)
}

type NotifyTransport struct {
	notifier notification.EmailClient
	match    string
}

func (t *NotifyTransport) RoundTrip(request *http.Request) (*http.Response, error) {

	reqBody, _ := io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	req := &SoapEnvelope{}
	_ = xml.Unmarshal(reqBody, req)

	soapServiceName := req.Body.Service.XMLName.Local
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return response, err
	}

	if response.StatusCode >= 400 {
		t.notify(response, soapServiceName)
	}

	return response, err
}

func (t *NotifyTransport) notify(response *http.Response, soapService string) {

	if t.match != soapService {
		return
	}

	respBody, _ := io.ReadAll(response.Body)
	response.Body = io.NopCloser(bytes.NewBuffer(respBody))

	// format soap response
	soapBody := xmlfmt.FormatXML(string(respBody), "", "  ")
	// send notification
	msg := fmt.Sprintf("gICS responded with error code %d:\n\n%s", response.StatusCode, soapBody)

	t.notifier.Send(msg)

}

type SoapEnvelope struct {
	XMLName xml.Name
	Body    Body
}

type Service struct {
	XMLName xml.Name
}

type Body struct {
	XMLName xml.Name
	Service Service `xml:",any"`
}
