package log

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xsmartdns/xsmartdns/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

var defaultLogger *logrus.Logger

func Init(cfg *config.Log) {
	defaultLogger = logrus.New()
	defaultLogger.SetLevel(convertToLevel(cfg.Level))
	defaultLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.DateTime,
	})
	if len(cfg.Filename) == 0 {
		defaultLogger.Out = os.Stdout
	} else {
		defaultLogger.Out = &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		}
	}
}

func convertToLevel(level string) logrus.Level {
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	}
	return logrus.InfoLevel
}

func Tracef(format string, args ...interface{}) {
	defaultLogger.Tracef(format, args...)
}
func Debuf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

func IsLevelEnabled(level logrus.Level) bool {
	return defaultLogger.IsLevelEnabled(level)
}
