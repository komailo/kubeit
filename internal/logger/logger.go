package logger

import (
	"github.com/sirupsen/logrus"
)

// Logger is the global logger instance
var Logger = logrus.New()

func init() {
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// SetLevelFromVerbosity sets the logger level based on how many times
// "-v" was specified (e.g., 0 for warn, 1 for info, 2 for debug, 3+ for trace).
func SetLevelFromVerbosity(v int) {
	switch v {
	case 0:
		Logger.SetLevel(logrus.InfoLevel)
	case 1:
		Logger.SetLevel(logrus.DebugLevel)
	default:
		Logger.SetLevel(logrus.TraceLevel) // For 3 or more -v's
	}
}

// need to find a better way to do these mapping. But later.
func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Trace(args ...interface{}) {
	Logger.Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	Logger.Tracef(format, args...)
}
