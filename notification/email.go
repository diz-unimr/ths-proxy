package notification

import (
	"fmt"
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/wneessen/go-mail"
	"log/slog"
	"strings"
	"time"
)

type Details func() string

type EmailClient interface {
	Send(subject, msg string, details string)
}

type DetailLevel int

const (
	LevelInfo DetailLevel = iota
	LevelDebug
)

var detailLevels = map[DetailLevel]string{
	LevelInfo:  "info",
	LevelDebug: "debug",
}

func (dl DetailLevel) String() string {
	return detailLevels[dl]
}

type emailClient struct {
	Sender     string
	Recipients map[DetailLevel][]string
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

	recp := map[DetailLevel][]string{
		LevelInfo:  strings.Split(config.Recipients.Info, ","),
		LevelDebug: strings.Split(config.Recipients.Debug, ","),
	}

	return &emailClient{
		Sender:     config.Sender,
		Recipients: recp,
		client:     client,
		throttle:   time.Tick(1 * time.Second),
	}
}

func (c *emailClient) Send(subject, msg string, details string) {

	for level, recp := range c.Recipients {
		if level == LevelDebug {
			c.sendTo(recp, subject, fmt.Sprintf("%s\n\nDetails:\n%s", msg, details))

		} else {
			c.sendTo(recp, subject, msg)
		}
	}
	slog.Debug("Notification sent", "type", "email")
}

func (c *emailClient) sendTo(recp []string, subject string, body string) {
	// throttle messages
	<-c.throttle

	message := mail.NewMsg()
	if err := message.From(c.Sender); err != nil {
		slog.Error("Failed to set FROM address.", "sender", c.Sender, "error", err)
		return
	}
	if err := message.To(recp...); err != nil {
		slog.Error("Failed to set TO address.", "recipients", c.Recipients, "error", err)
		return
	}
	message.Subject(subject)
	message.SetBodyString(mail.TypeTextPlain, body)

	if err := c.client.DialAndSend(message); err != nil {
		slog.Error("Failed to deliver E-Mail", "error", err)
		return
	}
	slog.Info("E-Mail notification successfully delivered")
}
