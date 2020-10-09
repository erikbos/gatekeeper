package shared

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is
type Logger struct {

	// Logging level of this logger
	Level string `yaml:"level"`

	// Filename to write log entries to
	Filename string `yaml:"filename"`

	// Maxsize per log file
	MaxSize int `yaml:"maxsixe"`

	//
	MaxBackups int `yaml:"maxbackups"`

	//
	MaxAge int `yaml:"maxage"`

	// logger is stored so we can use it to Rotate() it later on
	logger *lumberjack.Logger
}

// NewLogger returns a new logger
func NewLogger(config *Logger, applicationlog bool) *zap.Logger {

	// Open file write that can rotates for us
	config.logger = &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize, // megabytes
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,

		// Always log in UTC
		LocalTime: false,
	}
	writer := zapcore.AddSync(config.logger)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		MessageKey:     "msg",
		NameKey:        "logger",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if applicationlog {
		encoderConfig.LevelKey = "level"
		encoderConfig.StacktraceKey = "stacktrace"
		encoderConfig.StacktraceKey = "caller"
	}

	// Parse log level
	logLevel := zap.NewAtomicLevel()
	logLevel.UnmarshalText([]byte(config.Level))

	return zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, logLevel))
}

// Rotate closes and reopens logfile so old logfiles can be expired.
func (l *Logger) Rotate() {

	l.logger.Rotate()
}
