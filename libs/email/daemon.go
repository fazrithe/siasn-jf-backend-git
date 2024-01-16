package email

import (
	"crypto/tls"
	"github.com/if-itb/siasn-libs-backend/logutil"
	gomail "gopkg.in/mail.v2"
	"net"
	"strconv"
	"time"
)

const (
	SslModeUnencrypted = iota
	SslModeSsl
	SslModeStartTls
)

type Sender interface {
	Send(m *gomail.Message) error
}

// Daemon provides a queue and connection reuse.
// Email send requests are sent to SMTP server using the same connection, and after a set period of time
// that is configured by ConnExpireDuration, the connection is closed. The connection is reopened if there is another
// email send request.
//
// Because sending an email using the daemon is asynchronous, no error can be returned, therefore Send will
// always return nil error.
type Daemon struct {
	Host      string
	Port      int
	Username  string
	Password  string
	TlsConfig *tls.Config
	Logger    logutil.Logger
	sendCh    chan *gomail.Message
	Timeout   time.Duration
	SslMode   int
}

func NewDaemon(address string, username, password string, sslMode int) *Daemon {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	return &Daemon{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		sendCh:   make(chan *gomail.Message, 32),
		Timeout:  30 * time.Second,
		SslMode:  sslMode,
	}
}

// Start the daemon synchronously.
// It will return a channel that you can use to send emails through this daemon.
func (d *Daemon) Start() {
	dialer := gomail.NewDialer(d.Host, d.Port, d.Username, d.Password)
	switch d.SslMode {
	case SslModeSsl:
		dialer.SSL = true
	case SslModeStartTls:
		dialer.StartTLSPolicy = gomail.MandatoryStartTLS
	}
	dialer.TLSConfig = d.TlsConfig
	dialer.Timeout = d.Timeout

	var err error

	for {
		m, ok := <-d.sendCh
		if !ok {
			return
		}
		var s gomail.SendCloser
		if s, err = dialer.Dial(); err != nil {
			d.Logger.Warnf("cannot dial SMTP connection: %v", err)
			continue
		}
		if err = gomail.Send(s, m); err != nil {
			d.Logger.Warnf("cannot send email: %v", err)
			_ = s.Close()
			continue
		}
		_ = s.Close()
		d.Logger.Tracef("email successfully sent to %s", m.GetHeader("To"))
	}
}

// StartSender starts the daemon with a custom gomail.Sender.
// Unlike Start, because we do not know how to re build this custom gomail.Sender, connections are not closed and
// reopened.
func (d *Daemon) StartSender(s gomail.Sender) {
	d.sendCh = make(chan *gomail.Message)

	for {
		m, ok := <-d.sendCh
		if !ok {
			return
		}

		if err := gomail.Send(s, m); err != nil {
			d.Logger.Warnf("cannot send email: %v", err)
			break
		}
		d.Logger.Tracef("email successfully sent to %s", m.GetHeader("To"))
	}
}

// Send adds email to the queue.
// error will always be nil, but it can panic if this Daemon has been shutdown.
func (d *Daemon) Send(m *gomail.Message) error {
	d.sendCh <- m
	return nil
}

// Shutdown shuts down this daemon.
// This will also close the sendCh channel.
func (d *Daemon) Shutdown() {
	close(d.sendCh)
}
