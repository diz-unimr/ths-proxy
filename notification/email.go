package notification

import (
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/wneessen/go-mail"
	"log/slog"
	"strings"
	"time"
)

type EmailClient interface {
	Send(msg string)
}

type emailClient struct {
	Sender     string
	Recipients []string
	client     *mail.Client
	throttle   <-chan time.Time
}

func NewEmailClient(config config.Email) EmailClient {
	opts := []mail.Option{mail.WithTLSPolicy(mail.TLSOpportunistic)}

	if config.Smtp.Port != 0 {
		opts = append(opts, mail.WithPort(config.Smtp.Port))
	}
	if config.Smtp.User != "" {
		opts = append(opts, mail.WithUsername(config.Smtp.User))
	}
	if config.Smtp.Password != "" {
		opts = append(opts, mail.WithPassword(config.Smtp.Password))
	}
	if config.Smtp.Helo != "" {
		opts = append(opts, mail.WithHELO(config.Smtp.Helo))
	}

	client, err := mail.NewClient(config.Smtp.Server, opts...)
	if err != nil {
		slog.Error("Failed to create e-mail client", "error", err)
		return nil
	}

	return &emailClient{
		Sender:     config.Sender,
		Recipients: strings.Split(config.Recipients, ","),
		client:     client,
		throttle:   time.Tick(1 * time.Second),
	}
}

func (c *emailClient) Send(msg string) {

	// throttle messages
	<-c.throttle

	message := mail.NewMsg()
	if err := message.From(c.Sender); err != nil {
		slog.Error("Failed to set FROM address.", "sender", c.Sender, "error", err)
		return
	}
	if err := message.To(c.Recipients...); err != nil {
		slog.Error("Failed to set TO address.", "recipients", c.Recipients, "error", err)
		return
	}
	message.Subject("⚠️ gICS addConsent failed")
	message.SetBodyString(mail.TypeTextPlain, msg)

	if err := c.client.DialAndSend(message); err != nil {
		slog.Error("Failed to deliver E-Mail", "error", err)
		return
	}
	slog.Info("E-Mail notification successfully delivered")
}
