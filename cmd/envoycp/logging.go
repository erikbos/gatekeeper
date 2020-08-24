package main

import (
	log "github.com/sirupsen/logrus"
)

type logger struct{}

func (logger logger) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}
func (logger logger) Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}
func (logger logger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}
func (logger logger) Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}
