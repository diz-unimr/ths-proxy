package notification

import (
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/wneessen/go-mail"
	"log/slog"
	"strings"
	"time"
)

type EmailClient struct {
	Sender     string
	Recipients []string
	client     *mail.Client
	throttle   <-chan time.Time
}

func NewEmailClient(config config.Email) *EmailClient {
	client, err := mail.NewClient(config.Smtp.Server,
		mail.WithSMTPAuth(mail.SMTPAuthLogin), mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithUsername(config.Smtp.User), mail.WithPassword(config.Smtp.Password),
	)
	if err != nil {
		slog.Error("Failed to create e-mail client", "error", err)
		return nil
	}

	return &EmailClient{
		Sender:     config.Sender,
		Recipients: strings.Split(config.Recipients, ","),
		client:     client,
		throttle:   time.Tick(1 * time.Second),
	}
}

func (c *EmailClient) Send(msg string) {

	// throttle messages
	<-c.throttle

	message := mail.NewMsg()
	if err := message.From(c.Sender); err != nil {
		slog.Error("Failed to set FROM address.", "sender", c.Sender, "error", err)
	}
	if err := message.To(c.Recipients...); err != nil {
		slog.Error("Failed to set TO address.", "recipients", c.Recipients, "error", err)
	}
	message.Subject("⚠️ gICS AddConsent failed")
	message.SetBodyString(mail.TypeTextPlain, msg)

	if err := c.client.DialAndSend(message); err != nil {
		slog.Error("Failed to deliver E-Mail", "error", err)
		return
	}
	slog.Info("E-Mail notification successfully delivered")
}
