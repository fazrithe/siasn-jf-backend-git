package logutil

import (
	"fmt"
	"log"
	"os"
)

type StdLogger struct {
	stdOutLogger *log.Logger
	stdErrLogger *log.Logger
	Color        bool
	Prefix       string
}

func NewStdLogger(color bool, prefix string) *StdLogger {
	return &StdLogger{
		stdOutLogger: log.New(os.Stdout, "", log.LstdFlags),
		stdErrLogger: log.New(os.Stderr, "", log.LstdFlags),
		Color:        color,
		Prefix:       prefix,
	}
}

func (stdLogger *StdLogger) Trace(message string) {
	prefix := PrefixTrace
	if stdLogger.Color {
		prefix = PrefixTraceColor
	}

	stdLogger.printf("[%s] %s: %s", prefix, stdLogger.Prefix, message)
}

func (stdLogger *StdLogger) Tracef(format string, v ...interface{}) {
	stdLogger.Trace(fmt.Sprintf(format, v...))
}

func (stdLogger *StdLogger) Debug(message string) {
	prefix := PrefixDebug
	if stdLogger.Color {
		prefix = PrefixDebugColor
	}

	stdLogger.printf("[%s] %s: %s", prefix, stdLogger.Prefix, message)
}

func (stdLogger *StdLogger) Debugf(format string, v ...interface{}) {
	stdLogger.Debug(fmt.Sprintf(format, v...))
}

func (stdLogger *StdLogger) Info(message string) {
	prefix := PrefixInfo
	if stdLogger.Color {
		prefix = PrefixInfoColor
	}

	stdLogger.printf("[%s] %s: %s", prefix, stdLogger.Prefix, message)
}

func (stdLogger *StdLogger) Infof(format string, v ...interface{}) {
	stdLogger.Info(fmt.Sprintf(format, v...))
}

func (stdLogger *StdLogger) Warn(message string) {
	prefix := PrefixWarning
	if stdLogger.Color {
		prefix = PrefixWarningColor
	}

	stdLogger.stdErrLogger.Printf("[%s] %s: %s", prefix, stdLogger.Prefix, message)
}

func (stdLogger *StdLogger) Warnf(format string, v ...interface{}) {
	stdLogger.Warn(fmt.Sprintf(format, v...))
}

func (stdLogger *StdLogger) Error(message string) {
	prefix := PrefixFatal
	if stdLogger.Color {
		prefix = PrefixFatalColor
	}

	stdLogger.stdErrLogger.Printf("[%s] %s: %s", prefix, stdLogger.Prefix, message)
}

func (stdLogger *StdLogger) Errorf(format string, v ...interface{}) {
	stdLogger.Error(fmt.Sprintf(format, v...))
}

func (stdLogger *StdLogger) Access(message string) {
	prefix := PrefixAccess
	if stdLogger.Color {
		prefix = PrefixAccessColor
	}

	stdLogger.printf("[%s] %s: %s", prefix, stdLogger.Prefix, message)
}

func (stdLogger *StdLogger) printf(format string, v ...interface{}) {
	stdLogger.stdOutLogger.Printf(format, v...)
}
