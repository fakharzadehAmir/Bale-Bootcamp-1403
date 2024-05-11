package middleware

import "github.com/sirupsen/logrus"

// LogrusJaegerAdapter adapts a logrus.Logger to be used as a jaeger.Logger
type LogrusJaegerAdapter struct {
	logger *logrus.Logger
}

func (l *LogrusJaegerAdapter) Error(msg string) {
	l.logger.Error(msg)
}

func (l *LogrusJaegerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}
