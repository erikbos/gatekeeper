package main

import (
	"fmt"

	"go.uber.org/zap"
)

type cacheLogger struct {
	logger *zap.Logger
}

func newCacheLogger(logger *zap.Logger) cacheLogger {

	return cacheLogger{logger: logger}
}

func (cl cacheLogger) Infof(format string, args ...interface{}) {

	cl.logger.Info(fmt.Sprintf(format, args...))
}

func (cl cacheLogger) Warnf(format string, args ...interface{}) {

	cl.logger.Warn(fmt.Sprintf(format, args...))
}

func (cl cacheLogger) Errorf(format string, args ...interface{}) {

	cl.logger.Error(fmt.Sprintf(format, args...))
}

func (cl cacheLogger) Debugf(format string, args ...interface{}) {

	cl.logger.Debug(fmt.Sprintf(format, args...))
}
