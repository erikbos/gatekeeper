package shared

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger holds our log file configuration
type Logger struct {

	// Logging level of this logger
	Level string `yaml:"level"`

	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.
	Filename string `yaml:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `yaml:"maxsixe"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `yaml:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `yaml:"maxbackups"`

	// logger is stored so we can use it to Rotate() it later on
	logger *lumberjack.Logger
}

// NewLogger returns a new logger
func NewLogger(config *Logger) *zap.Logger {

	// Open file write that can rotates for us
	config.logger = &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,

		// Always log in UTC
		LocalTime: false,
	}
	writer := zapcore.AddSync(config.logger)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "t",
		MessageKey:     "m",
		NameKey:        "lg",
		LevelKey:       "l",
		StacktraceKey:  "s",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Parse log level
	logLevel := zap.NewAtomicLevel()
	if err := logLevel.UnmarshalText([]byte(config.Level)); err != nil {
		logLevel.SetLevel(zapcore.InfoLevel)
	}

	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, logLevel))
	logger.Info("Log opened", zap.String("loglevel", logLevel.String()))

	return logger
}

// Rotate closes and reopens logfile so old logfiles can be expired.
func (l *Logger) Rotate() {

	_ = l.logger.Rotate()
}
