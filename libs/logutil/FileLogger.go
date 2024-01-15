package logutil

import (
	"fmt"
	"log"
	"os"
	"path"
)

// FileLogger output logs into a file.
type FileLogger struct {
	logger *log.Logger
	Prefix string
}

// Creates a FileLogger that outputs logs to file pointed by filepath.
// If filepath does not exist, it will be created. If the directory does not exist,
func NewFileLogger(filepath, prefix string) *FileLogger {
	directory := path.Dir(filepath)
	if directory != "" {
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			panic(err)
		}
	}

	f, err := os.OpenFile(path.Clean(filepath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	return &FileLogger{
		logger: log.New(f, "", log.LstdFlags),
		Prefix: prefix,
	}
}

func (fileLogger *FileLogger) Trace(message string) {
	prefix := PrefixTrace
	fileLogger.printf("[%s] %s: %s", prefix, fileLogger.Prefix, message)
}

func (fileLogger *FileLogger) Tracef(format string, v ...interface{}) {
	fileLogger.Trace(fmt.Sprintf(format, v...))
}

func (fileLogger *FileLogger) Debug(message string) {
	prefix := PrefixDebug
	fileLogger.printf("[%s] %s: %s", prefix, fileLogger.Prefix, message)
}

func (fileLogger *FileLogger) Debugf(format string, v ...interface{}) {
	fileLogger.Debug(fmt.Sprintf(format, v...))
}

func (fileLogger *FileLogger) Info(message string) {
	prefix := PrefixInfo
	fileLogger.printf("[%s] %s: %s", prefix, fileLogger.Prefix, message)
}

func (fileLogger *FileLogger) Infof(format string, v ...interface{}) {
	fileLogger.Info(fmt.Sprintf(format, v...))
}

func (fileLogger *FileLogger) Warn(message string) {
	prefix := PrefixWarning
	fileLogger.printf("[%s] %s: %s", prefix, fileLogger.Prefix, message)
}

func (fileLogger *FileLogger) Warnf(format string, v ...interface{}) {
	fileLogger.Warn(fmt.Sprintf(format, v...))
}

func (fileLogger *FileLogger) Error(message string) {
	prefix := PrefixFatal
	fileLogger.printf("[%s] %s: %s", prefix, fileLogger.Prefix, message)
}

func (fileLogger *FileLogger) Errorf(format string, v ...interface{}) {
	fileLogger.Error(fmt.Sprintf(format, v...))
}

func (fileLogger *FileLogger) Access(message string) {
	prefix := PrefixAccess
	fileLogger.printf("[%s] %s: %s", prefix, fileLogger.Prefix, message)
}

func (fileLogger *FileLogger) printf(format string, v ...interface{}) {
	fileLogger.logger.Printf(format, v...)
}
