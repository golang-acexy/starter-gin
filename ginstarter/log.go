package ginstarter

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	level logrus.Level
}

func (l *logrusLogger) Write(p []byte) (n int, err error) {
	logger.Logrus().Logln(l.level, string(p))
	return len(p), nil
}
