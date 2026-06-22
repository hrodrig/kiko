// Package log provides leveled logging matching Microsoft LogLevel semantics.
//
// Levels (ascending severity):
//
//	Trace (0) — most verbose, diagnostic detail
//	Debug (1) — debugging info
//	Info  (2) — general operational messages
//	Warn  (3) — non-critical issues
//	Error (4) — runtime errors
//	Fatal (5) — critical failures, process exiting
//	Off   (6) — nothing logged
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Level is a logging severity level.
type Level int

const (
	Trace Level = iota // 0: most verbose
	Debug              // 1
	Info               // 2
	Warn               // 3
	Error              // 4
	Fatal              // 5: critical, process exiting
	Off                // 6: nothing logged
)

var levelNames = map[Level]string{
	Trace: "TRACE",
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
	Fatal: "FATAL",
}

// ParseLevel converts a string to Level.
// Accepts: trace, debug, info, warn, error, fatal, off (case-insensitive).
func ParseLevel(s string) (Level, error) {
	switch s {
	case "trace":
		return Trace, nil
	case "debug":
		return Debug, nil
	case "info":
		return Info, nil
	case "warn", "warning":
		return Warn, nil
	case "error":
		return Error, nil
	case "fatal", "critical":
		return Fatal, nil
	case "off", "none":
		return Off, nil
	default:
		return Info, fmt.Errorf("unknown log level: %q", s)
	}
}

// Logger is a leveled logger.
type Logger struct {
	level   Level
	logger  *log.Logger
	appName string
}

// New creates a Logger that writes to w.
// Default level is Info.
func New(w io.Writer, level Level) *Logger {
	if w == nil {
		w = os.Stderr
	}
	return &Logger{
		level:   level,
		logger:  log.New(w, "", 0),
		appName: "kiko",
	}
}

// WithAppName returns a copy of the Logger with a different app name.
func (l *Logger) WithAppName(name string) *Logger {
	return &Logger{
		level:   l.level,
		logger:  l.logger,
		appName: name,
	}
}

// SetLevel changes the minimum level.
func (l *Logger) SetLevel(level Level) { l.level = level }

// Level returns the current minimum level.
func (l *Logger) Level() Level { return l.level }

// LevelName returns the string representation of the current level (lowercase).
func (l *Logger) LevelName() string {
	return levelNames[l.level]
}

func (l *Logger) log(level Level, format string, args ...any) {
	if level < l.level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	ts := time.Now().UTC().Format(time.RFC3339)
	l.logger.Printf("%s  - %s - %-5s  %s", ts, l.appName, levelNames[level], msg)

	if level == Fatal {
		os.Exit(1)
	}
}

// Trace logs at Trace level.
func (l *Logger) Trace(format string, args ...any) { l.log(Trace, format, args...) }

// Debug logs at Debug level.
func (l *Logger) Debug(format string, args ...any) { l.log(Debug, format, args...) }

// Info logs at Info level.
func (l *Logger) Info(format string, args ...any) { l.log(Info, format, args...) }

// Warn logs at Warn level.
func (l *Logger) Warn(format string, args ...any) { l.log(Warn, format, args...) }

// Error logs at Error level.
func (l *Logger) Error(format string, args ...any) { l.log(Error, format, args...) }

// Fatal logs at Fatal level then exits with code 1.
func (l *Logger) Fatal(format string, args ...any) { l.log(Fatal, format, args...) }

// Default logger at Info level writing to stderr.
var Default = New(os.Stderr, Info)

// Convenience wrappers for Default logger.
func Tracef(format string, args ...any) { Default.log(Trace, format, args...) }
func Debugf(format string, args ...any) { Default.log(Debug, format, args...) }
func Infof(format string, args ...any)  { Default.log(Info, format, args...) }
func Warnf(format string, args ...any)  { Default.log(Warn, format, args...) }
func Errorf(format string, args ...any) { Default.log(Error, format, args...) }
func Fatalf(format string, args ...any) { Default.log(Fatal, format, args...) }
