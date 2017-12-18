// Package logger centralizes logging things in a way that gives similar output
// to Python tools.  For now, there is no filtering via log levels, and the
// output format is not yet customizable.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// TimeFormat exposes the standard time format string the default logger sets
const TimeFormat = "2006/01/02 15:04:05.000"

// LogLevel restricts log levels to a python-like set of numbers
type LogLevel int

// Python-like LogLevel constants
const (
	Debug LogLevel = 10
	Info  LogLevel = 20
	Warn  LogLevel = 30
	Err   LogLevel = 40
	Crit  LogLevel = 50
)

// String returns the standard text string for a given log level
func (l LogLevel) String() string {
	return logLevelStrings[l]
}

// Loggable defines a simple Log method loggers have to implement
type Loggable interface {
	Log(LogLevel, string)
}

var logLevelStrings = map[LogLevel]string{
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Err:   "ERROR",
	Crit:  "CRIT",
}

// SimpleLogger holds basic data to format log messages
type SimpleLogger struct {
	TimeFormat string
	AppName    string
	Output     io.Writer
}

// Logger wraps any loggable to add convenience methods for each log level:
// Debugf, Infof, etc.
type Logger struct {
	Loggable
}

// DefaultLogger gives an app semi-sane logging without creating and managing a
// custom type
var DefaultLogger = &Logger{
	&SimpleLogger{
		TimeFormat: TimeFormat,
		AppName:    filepath.Base(os.Args[0]),
		Output:     os.Stderr,
	},
}

// Log is the central logger for all helpers to use, implementing the Loggable interface
func (l *SimpleLogger) Log(level LogLevel, message string) {
	var timeString = time.Now().Format(l.TimeFormat)
	var output = fmt.Sprintf("%s - %s - %s - ", timeString, l.AppName, level)
	fmt.Fprintf(l.Output, output)
	fmt.Fprintln(l.Output, message)
}

// Debugf logs a debug-level message using the default logger
func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

// Debugf logs a debug-level message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(Debug, fmt.Sprintf(format, args...))
}

// Infof logs an info-level message using the default logger
func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

// Infof logs an info-level message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(Info, fmt.Sprintf(format, args...))
}

// Warnf logs a warn-level message using the default logger
func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}

// Warnf logs a warn-level message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Log(Warn, fmt.Sprintf(format, args...))
}

// Errorf logs an error-level message using the default logger
func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

// Errorf logs an error-level message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log(Err, fmt.Sprintf(format, args...))
}

// Criticalf logs a critical-level message using the default logger
func Criticalf(format string, args ...interface{}) {
	DefaultLogger.Criticalf(format, args...)
}

// Criticalf logs a critical-level message
func (l *Logger) Criticalf(format string, args ...interface{}) {
	l.Log(Crit, fmt.Sprintf(format, args...))
}

// Fatalf logs a critical-level message using the default logger, then exits
func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args...)
}

// Fatalf logs a critical-level message, then exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Log(Crit, fmt.Sprintf(format, args...))
	os.Exit(1)
}
