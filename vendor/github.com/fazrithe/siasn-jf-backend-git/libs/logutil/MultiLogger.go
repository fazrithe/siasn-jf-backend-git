package logutil

// MultiLogger allows you to log to multiple Logger at once.
type MultiLogger struct {
	loggers []Logger
}

// Create a new MultiLogger logging to each of the Logger given as the parameter.
func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{loggers: loggers}
}

func (multiLogger *MultiLogger) Trace(message string) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Trace(message)
	}
}

func (multiLogger *MultiLogger) Tracef(format string, v ...interface{}) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Tracef(format, v...)
	}
}

func (multiLogger *MultiLogger) Debug(message string) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Debug(message)
	}
}

func (multiLogger *MultiLogger) Debugf(format string, v ...interface{}) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Debugf(format, v...)
	}
}

func (multiLogger *MultiLogger) Info(message string) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Info(message)
	}
}

func (multiLogger *MultiLogger) Infof(format string, v ...interface{}) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Infof(format, v...)
	}
}

func (multiLogger *MultiLogger) Warn(message string) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Warn(message)
	}
}

func (multiLogger *MultiLogger) Warnf(format string, v ...interface{}) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Warnf(format, v...)
	}
}

func (multiLogger *MultiLogger) Error(message string) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Error(message)
	}
}

func (multiLogger *MultiLogger) Errorf(format string, v ...interface{}) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Errorf(format, v...)
	}
}

func (multiLogger *MultiLogger) Access(message string) {
	for i := 0; i < len(multiLogger.loggers); i++ {
		multiLogger.loggers[i].Access(message)
	}
}
