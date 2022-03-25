package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var GlobalLogger *Logger

// Logger : Logger type
type Logger struct {
	handle *zap.SugaredLogger
}

// Debugf : Log level type Debugf
func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.handle.Debugf(format, args...)
}

// Infof : Log level type Infof
func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.handle.Infof(format, args...)
}

// Errorf : Log level type Errorf
func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.handle.Errorf(format, args...)
}

// Fatalf : Log level type Fatalf
func (logger *Logger) Fatalf(format string, args ...interface{}) {
	logger.handle.Fatalf(format, args...)
}

// Panicf : Log level type Panicf
func (logger *Logger) Panicf(format string, args ...interface{}) {
	logger.handle.Panicf(format, args...)
}

// Debug : Log level type Debug
func (logger *Logger) Debug(args ...interface{}) {
	logger.handle.Debug(args...)
}

// Info : Log level type Info
func (logger *Logger) Info(args ...interface{}) {
	logger.handle.Info(args...)
}

// Warn : Log level type Warn
func (logger *Logger) Warn(args ...interface{}) {
	logger.handle.Warn(args...)
}

// Error : Log level type Error
func (logger *Logger) Error(args ...interface{}) {
	logger.handle.Error(args...)
}

// Fatal : Log level type Fatal
func (logger *Logger) Fatal(args ...interface{}) {
	logger.handle.Fatal(args...)
}

// Panic : Log level type Panic
func (logger *Logger) Panic(args ...interface{}) {
	logger.handle.Panic(args...)
}

// NewLogger creates the new logger object.
func NewLogger(logLevel string) *Logger {
	logLevelMap := map[string]zapcore.Level{
		"DEBUG": zapcore.DebugLevel,
		"INFO":  zapcore.InfoLevel,
		"ERROR": zapcore.ErrorLevel,
		"WARN":  zapcore.WarnLevel,
		"FATAL": zapcore.FatalLevel,
		"PANIC": zapcore.PanicLevel}

	logLvl := logLevelMap[logLevel]

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLvl),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()

	defer logger.Sync()

	return &Logger{logger}
}
