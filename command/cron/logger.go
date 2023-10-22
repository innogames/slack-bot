package cron

import log "github.com/sirupsen/logrus"

// special logger is needed for the cron lib...
type cronLogger struct {
	logger *log.Logger
}

func newCronLogger() cronLogger {
	return cronLogger{log.StandardLogger()}
}

func (l cronLogger) Info(msg string, keysAndValues ...any) {
	data := append([]any{"Cron", msg}, keysAndValues...)
	l.logger.Debug(data...)
}

func (l cronLogger) Error(err error, msg string, keysAndValues ...any) {
	data := append([]any{"Cron", err, msg}, keysAndValues...)
	l.logger.Error(data...)
}
