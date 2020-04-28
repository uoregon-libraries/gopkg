// Package logger centralizes logging things in a way that gives similar output
// to Python tools.  For now, there is no filtering via log levels, and the
// output format is not yet customizable.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TimeFormat exposes the standard time format string the default logger sets
const TimeFormat = "2006/01/02 15:04:05.000"

// LogLevel restricts log levels to a python-like set of numbers
type LogLevel int

// Python-like LogLevel constants
const (
	Invalid LogLevel = 0
	Debug   LogLevel = 10
	Info    LogLevel = 20
	Warn    LogLevel = 30
	Err     LogLevel = 40
	Crit    LogLevel = 50
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

// LogLevelFromString returns the LogLevel for a given human-readable string,
// or else Invalid if the string doesn't map to one of our log levels
func LogLevelFromString(s string) LogLevel {
	for level, str := range logLevelStrings {
		if str == s {
			return level
		}
	}

	return Invalid
}

// SimpleLogger holds basic data to format log messages
type SimpleLogger struct {
	TimeFormat string
	AppName    string
	Output     io.Writer
	LogWriter  func(level LogLevel, message string)
}

// Logger wraps any loggable to add convenience methods for each log level:
// Debugf, Infof, etc.
type Logger struct {
	Loggable
}

// A LeveledLogger wraps SimpleLogger to filter by log level
type LeveledLogger struct {
	*SimpleLogger
	Level LogLevel
}

// Log prints the message via SimpleLogger if its level is at or above
// LeveledLogger's log level
func (ll *LeveledLogger) Log(level LogLevel, message string) {
	if level >= ll.Level {
		ll.SimpleLogger.Log(level, message)
	}
}

func standardSimpleLogger() *SimpleLogger {
	var s = &SimpleLogger{
		TimeFormat: TimeFormat,
		Output:     os.Stderr,
	}
	s.LogWriter = s.DefaultLog

	return s
}

// Let's make sure if there's a problem pulling the app name, we know right at
// the beginning of the app
var defaultName = filepath.Base(os.Args[0])

// New returns an appropriate Logger that filters logs which are less
// important than the given log level.  If log level "DEBUG" is chosen, nothing
// is filtered.
func New(level LogLevel, structured bool) *Logger {
	return Named(defaultName, level, structured)
}

// Named returns a logger using the given name instead of defaulting to the
// application's command-line name
func Named(appName string, level LogLevel, structured bool) *Logger {
	var sl = standardSimpleLogger()
	sl.AppName = appName
	if structured {
		sl.LogWriter = sl.StructuredLog
	}
	if level <= Debug {
		return &Logger{sl}
	}

	return &Logger{&LeveledLogger{sl, level}}
}

// Log delegates to the LogWriter to format the message
func (l *SimpleLogger) Log(level LogLevel, message string) {
	l.LogWriter(level, message)
}

// DefaultLog is the default centralized logger for all helpers to use,
// implementing the Loggable interface
func (l *SimpleLogger) DefaultLog(level LogLevel, message string) {
	var timeString = time.Now().Format(l.TimeFormat)
	var output = fmt.Sprintf("%s - %s - %s - ", timeString, l.AppName, level)
	fmt.Fprintln(l.Output, output+message)
}

// esc escapes backslashes and quotes
func esc(s string) string {
	s = strings.Replace(s, `\`, `\\`, -1)
	s = strings.Replace(s, `"`, `\"`, -1)
	s = strings.Replace(s, "\r\n", "<CRLF>", -1)
	s = strings.Replace(s, "\n", "<CR>", -1)
	s = strings.Replace(s, "\r", "<LF>", -1)
	return s
}

// StructuredLog is an outputter that just prints key-value pairs in a way
// that's more machine-readable but still mostly human-friendly
func (l *SimpleLogger) StructuredLog(level LogLevel, message string) {
	var parts = [][2]string{
		{"time", time.Now().Format(l.TimeFormat)},
		{"app", l.AppName},
		{"level", level.String()},
		{"message", message},
	}

	var outputParts []string
	for _, part := range parts {
		outputParts = append(outputParts, esc(part[0])+`="`+esc(part[1])+`"`)
	}
	fmt.Fprintln(l.Output, strings.Join(outputParts, " "))
}

// Debugf logs a debug-level message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(Debug, fmt.Sprintf(format, args...))
}

// Infof logs an info-level message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(Info, fmt.Sprintf(format, args...))
}

// Warnf logs a warn-level message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Log(Warn, fmt.Sprintf(format, args...))
}

// Errorf logs an error-level message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log(Err, fmt.Sprintf(format, args...))
}

// Criticalf logs a critical-level message
func (l *Logger) Criticalf(format string, args ...interface{}) {
	l.Log(Crit, fmt.Sprintf(format, args...))
}

// Fatalf logs a critical-level message, then exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Log(Crit, fmt.Sprintf(format, args...))
	os.Exit(1)
}
