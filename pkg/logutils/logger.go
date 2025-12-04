package logutils

import (
	"github.com/sirupsen/logrus"
)

type Fields = logrus.Fields
type Entry = logrus.Entry

// Info logs a message at level Info
func Info(args ...any) {
	logrus.Info(args...)
}

// Infof logs a formatted message at level Info
func Infof(format string, args ...any) {
	logrus.Infof(format, args...)
}

// Debug logs a message at level Debug
func Debug(args ...any) {
	logrus.Debug(args...)
}

// Debugf logs a formatted message at level Debug
func Debugf(format string, args ...any) {
	logrus.Debugf(format, args...)
}

// Warn logs a message at level Warn
func Warn(args ...any) {
	logrus.Warn(args...)
}

// Warnf logs a formatted message at level Warn
func Warnf(format string, args ...any) {
	logrus.Warnf(format, args...)
}

// Error logs a message at level Error
func Error(args ...any) {
	logrus.Error(args...)
}

// Errorf logs a formatted message at level Error
func Errorf(format string, args ...any) {
	logrus.Errorf(format, args...)
}

// Fatal logs a message at level Fatal and then exits
func Fatal(args ...any) {
	logrus.Fatal(args...)
}

// Fatalf logs a formatted message at level Fatal and then exits
func Fatalf(format string, args ...any) {
	logrus.Fatalf(format, args...)
}

// WithField creates an entry with a single field
func WithField(key string, value any) *Entry {
	return logrus.WithField(key, value)
}

// WithFields creates an entry with multiple fields
func WithFields(fields Fields) *Entry {
	return logrus.WithFields(fields)
}

// WithError creates an entry with an error field
func WithError(err error) *Entry {
	return logrus.WithError(err)
}
