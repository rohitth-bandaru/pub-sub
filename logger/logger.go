package logger

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Fields type alias for logrus.Fields
type Fields = logrus.Fields

// Logger interface abstracts logging operations
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger
}

// LogrusLogger implements Logger interface using logrus
type LogrusLogger struct {
	entry *logrus.Entry
}

// NewLogger creates a new logger instance
func NewLogger(level, format string) Logger {
	// Set log level
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			ForceColors:     true,
			TimestampFormat: time.RFC3339,
		})
	}

	return &LogrusLogger{
		entry: logrus.NewEntry(logrus.StandardLogger()),
	}
}

// Debug logs debug level message
func (l *LogrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Info logs info level message
func (l *LogrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Warn logs warn level message
func (l *LogrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Error logs error level message
func (l *LogrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Fatal logs fatal level message and exits
func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Debugf logs formatted debug level message
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Infof logs formatted info level message
func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Warnf logs formatted warn level message
func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// Errorf logs formatted error level message
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Fatalf logs formatted fatal level message and exits
func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// WithField adds a field to the logger
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusLogger{
		entry: l.entry.WithField(key, value),
	}
}

// WithFields adds multiple fields to the logger
func (l *LogrusLogger) WithFields(fields Fields) Logger {
	return &LogrusLogger{
		entry: l.entry.WithFields(logrus.Fields(fields)),
	}
}

// WithError adds an error to the logger
func (l *LogrusLogger) WithError(err error) Logger {
	return &LogrusLogger{
		entry: l.entry.WithError(err),
	}
}
