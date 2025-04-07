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
		HostAddress:       "hex.localhost",
	})

	if err := server.Start(); err != nil {
		fmt.Println(err)
	}

	hostAddress, portNumber := "hex.localhost", server.PortNumber()

	c := config.Email{
		Sender:     "test@localhost",
		Recipients: "test@localhost",
		Smtp: config.Smtp{
			Server: hostAddress,
			Port:   portNumber,
		},
	}
	client := NewEmailClient(c)

	msg := "TEST"
	client.Send(msg)

	assert.NotEmpty(t, server.Messages())
	for _, m := range server.Messages() {
		fmt.Println(m)
		assert.Equal(t, msg, m.DataResponse())
	}

	if err := server.Stop(); err != nil {
		fmt.Println(err)
	}
}
