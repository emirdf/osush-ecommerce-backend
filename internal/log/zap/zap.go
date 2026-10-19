package zap

import (
	"fmt"
	"os"
	"osush/internal/log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ log.Logger = &Logger{}

type Logger struct {
	L *zap.Logger
}

func NewLogger(env string) (*Logger, func(), error) {
	var config zap.Config

	if env == "development" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.LevelKey = "severity"
		config.EncoderConfig.MessageKey = "message"
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	if env != "development" {
		if err := os.MkdirAll("./logs", 0755); err == nil {
			lumberJackLogger := &lumberjack.Logger{
				Filename:   "./logs/app.log",
				MaxSize:    100,
				MaxBackups: 7,
				MaxAge:     30,
				Compress:   true,
			}

			filecore := zapcore.NewCore(
				zapcore.NewJSONEncoder(config.EncoderConfig),
				zapcore.AddSync(lumberJackLogger),
				zapcore.InfoLevel,
			)

			consoleCore := zapcore.NewCore(
				zapcore.NewJSONEncoder(config.EncoderConfig),
				zapcore.Lock(os.Stdout),
				zapcore.InfoLevel,
			)

			core := zapcore.NewTee(filecore, consoleCore)

			z := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))
			z = z.With(
				zap.String("service", "service-records-api"),
				zap.String("env", env),
			)

			l := &Logger{L: z}
			cleanup := func() { _ = l.L.Sync() }
			return l, cleanup, nil
		}
	}

	z, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.FatalLevel),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	z = z.With(
		zap.String("service", "service-records-api"),
		zap.String("env", env))

	l := &Logger{
		L: z,
	}

	cleanup := func() {
		_ = l.L.Sync()
	}

	return l, cleanup, nil
}

func (l *Logger) Trace(msg string, fields ...log.Field) {
	if ce := l.L.Check(zap.DebugLevel, msg); ce != nil {
		ce.Write(zapifyFields(fields...)...)
	}
}

func (l *Logger) Tracef(msg string, args ...interface{}) {
	if ce := l.L.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

func (l *Logger) Debug(msg string, fields ...log.Field) {
	if ce := l.L.Check(zap.DebugLevel, msg); ce != nil {
		ce.Write(zapifyFields(fields...)...)
	}
}

func (l *Logger) Debugf(msg string, args ...interface{}) {
	if ce := l.L.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

func (l *Logger) Info(msg string, fields ...log.Field) {
	if ce := l.L.Check(zap.InfoLevel, msg); ce != nil {
		ce.Write(zapifyFields(fields...)...)
	}
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	if ce := l.L.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

func (l *Logger) Warn(msg string, fields ...log.Field) {
	if ce := l.L.Check(zap.WarnLevel, msg); ce != nil {
		ce.Write(zapifyFields(fields...)...)
	}
}

func (l *Logger) Warnf(msg string, args ...interface{}) {
	if ce := l.L.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

func (l *Logger) Error(msg string, fields ...log.Field) {
	if ce := l.L.Check(zap.ErrorLevel, msg); ce != nil {
		ce.Write(zapifyFields(fields...)...)
	}
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	if ce := l.L.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

func (l *Logger) Fatal(msg string, fields ...log.Field) {
	if ce := l.L.Check(zap.FatalLevel, msg); ce != nil {
		ce.Write(zapifyFields(fields...)...)
	}
}

func (l *Logger) Fatalf(msg string, args ...interface{}) {
	if ce := l.L.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

func (l *Logger) With(fields ...log.Field) *Logger {
	return &Logger{
		L: l.L.With(zapifyFields(fields...)...),
	}
}
