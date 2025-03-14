package zap

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogWrapper func(level log.Level, keyvals ...any) error

func (f LogWrapper) Log(level log.Level, keyvals ...any) error {
	return f(level, keyvals...)
}

type Option struct {
	// Level is a logging priority (refer: zap/zapcore/level.go).
	// The default is 0 (info).
	// If Level sets to -1 (debug) output logs to the console.
	//
	// Choices: -1: debug, 0: info, 1: warn, 2: error, 3: dPanic, 4: panic, 5: fatal.
	// The default is info.
	Level zapcore.Level

	// FilePath is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	FilePath string

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated.
	// It defaults to 100 MB.
	MaxSize int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int

	// Compress determines if the rotated log files should be compressed using gzip.
	// The default is not to perform compression.
	Compress bool
}

var DefaultZapOption = Option{
	Level: zapcore.InfoLevel,
}

// NewLogger a wrapper for zap logger to implement log.Logger interface which built-in go-kratos.
func NewLogger(opts ...Option) (log.Logger, error) {
	var option Option
	if len(opts) == 0 {
		option = DefaultZapOption
	} else {
		option = opts[0]
	}

	var logger *zap.Logger
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "" // remove default time key

	// file encoder
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   option.FilePath,
		MaxSize:    option.MaxAge,
		MaxBackups: option.MaxBackups,
		MaxAge:     option.MaxAge,
		Compress:   option.Compress,
	})
	writeSyncer := zapcore.NewMultiWriteSyncer(fileWriteSyncer)
	fileCore := zapcore.NewCore(fileEncoder, writeSyncer, option.Level)
	// console encoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleDebugging, zapcore.DebugLevel)
	if option.Level == zapcore.DebugLevel {
		logger = zap.New(zapcore.NewTee(consoleCore, fileCore), zap.AddCaller())
	} else {
		logger = zap.New(zapcore.NewTee(fileCore), zap.AddCaller())
	}

	return LogWrapper(func(level log.Level, keyvals ...any) error {
		zapLevel := zap.DebugLevel
		switch level {
		case log.LevelDebug:
			zapLevel = zap.DebugLevel
		case log.LevelInfo:
			zapLevel = zap.InfoLevel
		case log.LevelWarn:
			zapLevel = zap.WarnLevel
		case log.LevelError:
			zapLevel = zap.ErrorLevel
		case log.LevelFatal:
			zapLevel = zap.FatalLevel
		}
		var fields []zap.Field
		for i := 0; i < len(keyvals); i += 2 {
			fields = append(fields, zap.String(fmt.Sprintf("%v", keyvals[i]), fmt.Sprintf("%v", keyvals[i+1])))
		}
		logger.Log(zapLevel, "", fields...)
		return nil
	}), nil
}
