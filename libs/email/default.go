package email

import (
	"crypto/tls"
	gomail "gopkg.in/mail.v2"
)

// DefaultSender sends email by making connection to SMTP server for every time a send request is made.
// It does not implement and queueing strategy at all. Read also Daemon.
type DefaultSender struct {
	Host      string
	Port      int
	Username  string
	Password  string
	TlsConfig *tls.Config
	SslMode   int
}

func (d *DefaultSender) Send(m *gomail.Message) error {
	dl := gomail.NewDialer(d.Host, d.Port, d.Username, d.Password)
	switch d.SslMode {
	case SslModeSsl:
		dl.SSL = true
	case SslModeStartTls:
		dl.StartTLSPolicy = gomail.MandatoryStartTLS
	}
	dl.TLSConfig = d.TlsConfig
	return dl.DialAndSend(m)
}
