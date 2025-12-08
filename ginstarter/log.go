package ginstarter

import (
	"strings"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	level logrus.Level
}

func (l *logrusLogger) Write(p []byte) (n int, err error) {
	str := string(p)
	str = strings.TrimRight(str, "\r\n")
	logger.Logrus().Log(l.level, str)
	return len(p), nil
}
