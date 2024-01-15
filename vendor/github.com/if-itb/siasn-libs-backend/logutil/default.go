package logutil

import (
	"os/exec"
	"strconv"
	"strings"
)

var defaultLogger Logger = NewStdLogger(IsSupportColor(), "")

const (
	PrefixTrace        = "TRACE "
	PrefixTraceColor   = "\x1b[37;1mTRACE \x1b[0m"
	PrefixDebug        = "DEBUG "
	PrefixDebugColor   = "\x1b[34;1mDEBUG \x1b[0m"
	PrefixInfo         = "INFO  "
	PrefixInfoColor    = "\x1b[36;1mINFO  \x1b[0m"
	PrefixWarning      = "WARN  "
	PrefixWarningColor = "\x1b[33;1mWARN  \x1b[0m"
	PrefixFatal        = "ERROR "
	PrefixFatalColor   = "\x1b[31;1mERROR \x1b[0m"
	PrefixAccess       = "ACCESS"
	PrefixAccessColor  = "\x1b[35;1mACCESS\x1b[0m"
)

// Checks whether the terminal supports color.
func IsSupportColor() bool {
	output, err := exec.Command("tput", "colors").Output()
	if err != nil {
		return false
	}

	colors, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil || colors < 8 {
		return false
	}

	return true
}

// Replaces the default logger with new logger.
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// Outputs a trace message with the default logger.
func Trace(message string) {
	defaultLogger.Trace(message)
}

// Outputs a trace formatted message with the default logger.
func Tracef(format string, v ...interface{}) {
	defaultLogger.Tracef(format, v...)
}

// Outputs a debug message with the default logger.
func Debug(message string) {
	defaultLogger.Debug(message)
}

// Outputs a debug formatted message with the default logger.
func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

// Outputs a info message with the default logger.
func Info(message string) {
	defaultLogger.Info(message)
}

// Outputs a info formatted message with the default logger.
func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

// Outputs a warning message with the default logger.
func Warn(message string) {
	defaultLogger.Warn(message)
}

// Outputs a warning formatted message with the default logger.
func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

// Outputs an error message with the default logger.
func Error(message string) {
	defaultLogger.Error(message)
}

// Outputs an error formatted message with the default logger.
func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

// Outputs an access message with the default logger.
func Access(message string) {
	defaultLogger.Access(message)
}
