package ginmodule

import "github.com/sirupsen/logrus"

type logrusLogger struct {
	log   *logrus.Logger
	level logrus.Level
}

func (l *logrusLogger) Write(p []byte) (n int, err error) {
	l.log.Log(l.level, string(p))
	return len(p), nil
}
