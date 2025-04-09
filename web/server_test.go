package web

import (
	"fmt"
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func NewTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

type MailClientMock struct {
	received string
}

func (c *MailClientMock) Send(_, msg string, _ func() string) {
	c.received = msg
}

type RouteTestCase struct {
	name     string
	method   string
	endpoint string
	body     io.Reader
	expPath  string
	expBody  string
}

type NotificationTestCase struct {
	name         string
	expectNotify bool
	mockStatus   string
	serviceName  string
}

func TestHandlers(t *testing.T) {
	// dummy gics endpoint returns request body and path for verification (in header: Request-Path)
	gics := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Request-Path", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprint(w, string(body))
	}))
	// setup config
	c := config.AppConfig{
		App: config.App{},
		Gics: config.Gics{
			BaseUrl: gics.URL,
		},
	}

	s := NewServer(c)
	s.client = http.DefaultClient
	defer gics.Close()

	origBody := `<consent></consent>`
	modBody := `<notificationClientID>gICS_Soap</notificationClientID><consent></consent>`

	cases := []RouteTestCase{
		{
			name: "Rewritten", method: "POST", endpoint: "/gics/gicsService",
			body: strings.NewReader(origBody), expPath: "/gics/gicsServiceWithNotification", expBody: modBody,
		},
		{
			name: "NoRoute_Method", method: "GET", endpoint: "/gics/gicsService",
			body: strings.NewReader(origBody), expPath: "/gics/gicsService", expBody: origBody,
		},
		{
			name: "NoRoute_Endpoint", method: "POST", endpoint: "/gics/gicsServiceWithNotification",
			body: strings.NewReader(origBody), expPath: "/gics/gicsServiceWithNotification", expBody: origBody,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testRoute(t, s, c.method, c.endpoint, c.body, c.expPath, c.expBody)
		})
	}
}

func testRoute(t *testing.T, s *Server, method, endpoint string, body io.Reader, expectedPath, expectedBody string) {
	// setup routes
	r := s.setupRouter()

	req, _ := http.NewRequest(method, endpoint, body)
	w := NewTestResponseRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, expectedPath, w.Header().Get("Request-Path"))
	assert.Equal(t, expectedBody, w.Body.String())
}

func TestNotification(t *testing.T) {

	// dummy gics endpoint
	gics := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResult, _ := strconv.Atoi(r.Header.Get("Mock-Status"))
		w.WriteHeader(mockResult)
		//_, _ = fmt.Fprintf(w, soapErrResponse)
	}))
	defer gics.Close()

	// setup config
	c := config.AppConfig{
		App: config.App{},
		Gics: config.Gics{
			BaseUrl: gics.URL,
		},
	}

	s := NewServer(c)

	cases := []NotificationTestCase{
		{
			name:         "SuccessResponse",
			mockStatus:   "200",
			expectNotify: false,
			serviceName:  "addConsent",
		},
		{
			name:         "Error_Response",
			mockStatus:   "400",
			expectNotify: true,
			serviceName:  "addConsent",
		},
		{
			name:         "Error_Response_No_Match",
			mockStatus:   "200",
			expectNotify: false,
			serviceName:  "invalid",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testNotification(t, s, c.expectNotify, c.serviceName, c.mockStatus)
		})
	}
}

func testNotification(t *testing.T, s *Server, expectNotify bool, serviceName, mockStatus string) {

	// setup test notification client
	mailMock := &MailClientMock{}
	s.proxy.Transport = &NotifyTransport{
		notifier: mailMock,
		match:    serviceName,
	}
	// setup routes
	r := s.setupRouter()

	reqBody := `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cm2="http://cm2.ttp.ganimed.icmvc.emau.org/">
				  <soapenv:Header/>
				  <soapenv:Body>
				    <cm2:addConsent>
                      <consent></consent>
					</cm2:addConsent>
				</soapenv:Body>
			</soapenv:Envelope>`

	req, _ := http.NewRequest(http.MethodPost, "/gics/gicsService", strings.NewReader(reqBody))
	req.Header.Set("Mock-Status", mockStatus)
	w := NewTestResponseRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, expectNotify, mailMock.received != "")
}
