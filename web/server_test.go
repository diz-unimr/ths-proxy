package web

import (
	"bytes"
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ClientMock struct {
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	responseBody := io.NopCloser(bytes.NewReader([]byte(
		`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
           <soap:Body>
             <ns2:addConsentResponse xmlns:ns2="http://cm2.ttp.ganimed.icmvc.emau.org/"/>
           </soap:Body>
         </soap:Envelope>`)))
	return &http.Response{
		StatusCode: 200,
		Header:     map[string][]string{"Content-Type": {"text/xml"}},
		Body:       responseBody}, nil
}

func TestHandleSoap(t *testing.T) {
	// setup config
	c := config.AppConfig{
		App: config.App{},
		Gics: config.Gics{
			BaseUrl: "",
		},
	}

	s := NewServer(c)
	s.client = &ClientMock{}

	body := `<consent></consent>`
	reqBody := []byte(body)
	statusCode := 200

	testRoute(t, s, "POST", "/gics/gicsService", bytes.NewBuffer(reqBody), statusCode)
}

func testRoute(t *testing.T, s *Server, method, endpoint string, body io.Reader, returnCode int) {
	r := s.setupRouter()

	req, _ := http.NewRequest(method, endpoint, body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, returnCode, w.Code)
}
