package notification

import (
	"fmt"
	"testing"

	"github.com/diz-unimr/ths-proxy/pkg/config"
	"github.com/diz-unimr/ths-proxy/pkg/consent"
	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/wneessen/go-mail"
)

func testSend(t *testing.T, doc *consent.Document) {
	server := smtpmock.New(smtpmock.ConfigurationAttr{
		LogToStdout:       true,
		LogServerActivity: true,
	})

	if err := server.Start(); err != nil {
		fmt.Println(err)
	}

	hostAddress, portNumber := "127.0.0.1", server.PortNumber()

	c := config.Email{
		Sender: "test@localhost",
		Recipients: config.Recipients{
			Info:  "test@localhost",
			Debug: "test@localhost",
		},
		Smtp: config.Smtp{
			Server: hostAddress,
			Port:   portNumber,
			Helo:   "localhost",
		},
	}
	client := NewEmailClient(c)

	client.Send("Oops, something went wrong", "TEST", "Body", doc)

	messages := server.MessagesAndPurge()
	assert.Len(t, messages, 2)

	if err := server.Stop(); err != nil {
		fmt.Println(err)
	}
}

func TestSend(t *testing.T) {
	cases := []struct {
		name string
		doc  *consent.Document
	}{
		{name: "WithoutAttachment", doc: nil},
		{name: "WithAttachment", doc: createTestDocument()},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testSend(t, c.doc)
		})
	}
}

func createTestDocument() *consent.Document {
	return &consent.Document{
		Name: "test",
		Type: "application/pdf",
		Data: new("JVBERi0xLjIgCjkgMCBvYmoKPDwKPj4Kc3RyZWFtCkJULyA5IFRmKFRlc3QpJyBFVAplbmRzdHJlYW0KZW5kb2JqCjQgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCA1IDAgUgovQ29udGVudHMgOSAwIFIKPj4KZW5kb2JqCjUgMCBvYmoKPDwKL0tpZHMgWzQgMCBSIF0KL0NvdW50IDEKL1R5cGUgL1BhZ2VzCi9NZWRpYUJveCBbIDAgMCA5OSA5IF0KPj4KZW5kb2JqCjMgMCBvYmoKPDwKL1BhZ2VzIDUgMCBSCi9UeXBlIC9DYXRhbG9nCj4+CmVuZG9iagp0cmFpbGVyCjw8Ci9Sb290IDMgMCBSCj4+CiUlRU9G"),
	}
}

func TestAddAttachment(t *testing.T) {
	doc := consent.Document{
		Name: "test",
		Type: "application/pdf",
		Data: new("JVBERi0xLjIgCjkgMCBvYmoKPDwKPj4Kc3RyZWFtCkJULyA5IFRmKFRlc3QpJyBFVAplbmRzdHJlYW0KZW5kb2JqCjQgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCA1IDAgUgovQ29udGVudHMgOSAwIFIKPj4KZW5kb2JqCjUgMCBvYmoKPDwKL0tpZHMgWzQgMCBSIF0KL0NvdW50IDEKL1R5cGUgL1BhZ2VzCi9NZWRpYUJveCBbIDAgMCA5OSA5IF0KPj4KZW5kb2JqCjMgMCBvYmoKPDwKL1BhZ2VzIDUgMCBSCi9UeXBlIC9DYXRhbG9nCj4+CmVuZG9iagp0cmFpbGVyCjw8Ci9Sb290IDMgMCBSCj4+CiUlRU9G"),
	}

	msg := mail.NewMsg()

	addAttachment(msg, new(doc))

	assert.Equal(t, msg.GetAttachments()[0].ContentType.String(), "application/pdf")
}
