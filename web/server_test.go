package web

import (
	"bytes"
	"fmt"
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	soapResponse = `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
           <soap:Body>
             <ns2:addConsentResponse xmlns:ns2="http://cm2.ttp.ganimed.icmvc.emau.org/"/>
           </soap:Body>
         </soap:Envelope>`
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

type ClientMock struct {
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	responseBody := io.NopCloser(bytes.NewReader([]byte(soapResponse)))
	return &http.Response{
		StatusCode: 200,
		Header:     map[string][]string{"Content-Type": {"text/xml"}},
		Body:       responseBody,
	}, nil
}

type TestCase struct {
	name         string
	method       string
	endpoint     string
	body         io.Reader
	expectedCode int
	expectedBody string
}

func TestHandlers(t *testing.T) {
	// setup config
	c := config.AppConfig{
		App: config.App{},
		Gics: config.Gics{
			BaseUrl: "",
		},
	}

	s := NewServer(c)
	s.client = &ClientMock{}
	// dummy gics server to test relay
	gics := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Request relayed to unaltered")
	}))
	defer gics.Close()
	gicsUrl, _ := url.Parse(gics.URL)
	s.gicsUrl = gicsUrl

	body := `<consent></consent>`
	reqBody := []byte(body)

	cases := []TestCase{
		{
			name: "Rewritten", method: "POST", endpoint: "/gics/gicsService",
			body: bytes.NewBuffer(reqBody), expectedCode: 200, expectedBody: soapResponse,
		},
		{
			name: "NoRoute_Method", method: "GET", endpoint: "/gics/gicsService",
			body: bytes.NewBuffer(reqBody), expectedCode: 200, expectedBody: "Request relayed to unaltered",
		},
		{
			name: "NoRoute_Endpoint", method: "POST", endpoint: "/gics/gicsServiceWithNotification",
			body: bytes.NewBuffer(reqBody), expectedCode: 200, expectedBody: "Request relayed to unaltered",
		},
	}

	testRoute(t, s, "POST", "/gics/gicsService", bytes.NewBuffer(reqBody), http.StatusOK, soapResponse)
	testRoute(t, s, "GET", "/gics/gicsService", bytes.NewBuffer(reqBody), http.StatusOK, "Request relayed to unaltered")
	testRoute(t, s, "POST", "/gics/gicsServiceWithNotification", bytes.NewBuffer(reqBody), http.StatusOK, "Request relayed to unaltered")

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testRoute(t, s, c.method, c.endpoint, bytes.NewBuffer(reqBody), c.expectedCode, c.expectedBody)
		})
	}
}

func testRoute(t *testing.T, s *Server, method, endpoint string, body io.Reader, expectedCode int, expectedBody string) {
	// setup routes
	r := s.setupRouter()

	req, _ := http.NewRequest(method, endpoint, body)
	w := NewTestResponseRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, expectedCode, w.Code)
	rBody, _ := io.ReadAll(w.Body)
	assert.Equal(t, expectedBody, string(rBody))
}
