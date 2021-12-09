package logger

import (
	"time"

	logger "github.com/sirupsen/logrus"
)

func Init(json bool) {

	if json {
		logger.SetFormatter(&logger.JSONFormatter{})
	}
	logger.SetLevel(logger.DebugLevel)
}

func Info(args ...interface{}) {
	logger.Info(args)
}

func Trace(args ...interface{}) {
	logger.Trace(args[0:])
}

func Debug(args ...interface{}) {
	logger.Debug(args[0:])
}

func Warn(args ...interface{}) {
	logger.Warn(args[0:])
}

func Error(args ...interface{}) {
	logger.Error(args[0:])
}

func getLogTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
