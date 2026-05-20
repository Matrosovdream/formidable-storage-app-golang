package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger(cfg LogConfig) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05Z07:00"})
	if level, err := logrus.ParseLevel(cfg.Level); err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	return log
}
