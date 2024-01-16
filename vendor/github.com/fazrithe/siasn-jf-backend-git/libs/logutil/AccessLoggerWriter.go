package logutil

// AccessLoggerWriter implements io.Writer.
// It is designed to be used in conjunction with github.com/gorilla/handlers (handlers.LoggingHandler) to log
// API access activities. handlers.LoggingHandler requires an io.Writer. This writer will pass the written message
// to an AccessLogger.
type AccessLoggerWriter struct {
	Logger AccessLogger
}

func (a *AccessLoggerWriter) Write(p []byte) (n int, err error) {
	a.Logger.Access(string(p))
	return len(p), nil
}
