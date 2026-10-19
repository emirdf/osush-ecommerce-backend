package zap

import "osush/internal/log"

var _ log.Logger = &NopLogger{}

type NopLogger struct{}

func NewNop() log.Logger {
	return &NopLogger{}
}

func (NopLogger) Trace(msg string, fields ...log.Field) {}
func (NopLogger) Debug(msg string, fields ...log.Field) {}
func (NopLogger) Info(msg string, fields ...log.Field)  {}
func (NopLogger) Warn(msg string, fields ...log.Field)  {}
func (NopLogger) Error(msg string, fields ...log.Field) {}
func (NopLogger) Fatal(msg string, fields ...log.Field) {}

func (NopLogger) Tracef(format string, args ...interface{}) {}
func (NopLogger) Debugf(format string, args ...interface{}) {}
func (NopLogger) Infof(format string, args ...interface{})  {}
func (NopLogger) Warnf(format string, args ...interface{})  {}
func (NopLogger) Errorf(format string, args ...interface{}) {}
func (NopLogger) Fatalf(format string, args ...interface{}) {}
