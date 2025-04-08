package notification

import (
	"fmt"
	"github.com/diz-unimr/ths-proxy/config"
	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSend(t *testing.T) {
	server := smtpmock.New(smtpmock.ConfigurationAttr{
		LogToStdout:       true,
		LogServerActivity: true,
	})

	if err := server.Start(); err != nil {
		fmt.Println(err)
	}

	hostAddress, portNumber := "127.0.0.1", server.PortNumber()

	c := config.Email{
		Sender:     "test@localhost",
		Recipients: "test@localhost",
		Smtp: config.Smtp{
			Server: hostAddress,
			Port:   portNumber,
			Helo:   "localhost",
		},
	}
	client := NewEmailClient(c)

	client.Send("TEST")

	messages := server.MessagesAndPurge()
	assert.Len(t, messages, 1)

	if err := server.Stop(); err != nil {
		fmt.Println(err)
	}
}
