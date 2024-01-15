package logutil

type Logger interface {
	Trace(message string)
	Tracef(format string, v ...interface{})
	Debug(message string)
	Debugf(format string, v ...interface{})
	Info(message string)
	Infof(format string, v ...interface{})
	Warn(message string)
	Warnf(format string, v ...interface{})
	Error(message string)
	Errorf(format string, v ...interface{})

	AccessLogger
}

type AccessLogger interface {
	Access(message string)
}
