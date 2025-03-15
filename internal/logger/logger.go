package logger

import "github.com/sirupsen/logrus"

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func SetLevelFromVerbosity(v int) {
	switch {
	case v <= 0:
		log.SetLevel(logrus.WarnLevel)
	case v == 1:
		log.SetLevel(logrus.InfoLevel)
	case v == 2:
		log.SetLevel(logrus.DebugLevel)
	default:
		log.SetLevel(logrus.TraceLevel)
	}
}

var (
	Info   = log.Info
	Infof  = log.Infof
	Warn   = log.Warn
	Warnf  = log.Warnf
	Error  = log.Error
	Errorf = log.Errorf
	Fatal  = log.Fatal
	Fatalf = log.Fatalf
	Debug  = log.Debug
	Debugf = log.Debugf
	Trace  = log.Trace
	Tracef = log.Tracef
)
