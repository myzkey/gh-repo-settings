package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Level represents log level
type Level int

const (
	LevelQuiet Level = iota
	LevelNormal
	LevelVerbose
)

// Logger provides structured logging with levels
type Logger struct {
	level  Level
	out    io.Writer
	errOut io.Writer
}

// global logger instance
var defaultLogger = New(LevelNormal)

// New creates a new Logger
func New(level Level) *Logger {
	return &Logger{
		level:  level,
		out:    os.Stdout,
		errOut: os.Stderr,
	}
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.out = w
}

// SetErrorOutput sets the error output writer
func (l *Logger) SetErrorOutput(w io.Writer) {
	l.errOut = w
}

// Debug prints debug messages (only in verbose mode)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level >= LevelVerbose {
		gray := color.New(color.FgHiBlack).SprintFunc()
		fmt.Fprintf(l.out, gray("[DEBUG] "+format)+"\n", args...)
	}
}

// Info prints info messages (normal and verbose mode)
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		fmt.Fprintf(l.out, format+"\n", args...)
	}
}

// Success prints success messages with green checkmark
func (l *Logger) Success(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(l.out, green("✓")+" "+format+"\n", args...)
	}
}

// Warn prints warning messages (always shown except in quiet mode)
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Fprintf(l.errOut, yellow("⚠")+" "+format+"\n", args...)
	}
}

// Error prints error messages (always shown)
func (l *Logger) Error(format string, args ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Fprintf(l.errOut, red("✗")+" "+format+"\n", args...)
}

// Print prints messages without any formatting
func (l *Logger) Print(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		fmt.Fprintf(l.out, format, args...)
	}
}

// Println prints messages with newline
func (l *Logger) Println(args ...interface{}) {
	if l.level >= LevelNormal {
		fmt.Fprintln(l.out, args...)
	}
}

// Progress prints inline progress (no newline)
func (l *Logger) Progress(format string, args ...interface{}) {
	if l.level >= LevelNormal {
		fmt.Fprintf(l.out, format, args...)
	}
}

// ProgressDone completes a progress line with success
func (l *Logger) ProgressDone() {
	if l.level >= LevelNormal {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintln(l.out, green("✓"))
	}
}

// ProgressFail completes a progress line with failure
func (l *Logger) ProgressFail() {
	if l.level >= LevelNormal {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintln(l.out, red("✗"))
	}
}

// Global functions that use defaultLogger

// SetDefaultLevel sets the default logger level
func SetDefaultLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// Default returns the default logger
func Default() *Logger {
	return defaultLogger
}

// Debug logs debug message
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info logs info message
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Success logs success message
func Success(format string, args ...interface{}) {
	defaultLogger.Success(format, args...)
}

// Warn logs warning message
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Error logs error message
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}
