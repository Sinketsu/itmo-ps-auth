package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

func New(ctx string) *logrus.Entry {

	newLog := logrus.New()
	formatter := logrus.TextFormatter{
		TimestampFormat: time.RFC1123,
		FullTimestamp: true,
		DisableTimestamp: false,
	}
	newLog.SetFormatter(&formatter)
	newLog.SetOutput(os.Stdout)

	level, err := logrus.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	newLog.SetLevel(level)

	return newLog.WithField("context", ctx)
}
