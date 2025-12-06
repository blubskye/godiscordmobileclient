// Copyright (C) 2025 blubskye
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// Source code: https://github.com/blubskye/godiscordmobileclient

package debug

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents the logging level
type Level int

const (
	LevelOff Level = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

func (l Level) String() string {
	switch l {
	case LevelOff:
		return "OFF"
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// Logger provides debug logging with stack traces
type Logger struct {
	mu            sync.RWMutex
	level         Level
	output        io.Writer
	file          *os.File
	enableStack   bool
	stackDepth    int
	component     string
	showTimestamp bool
	showCaller    bool
}

var (
	// Default is the default global logger
	Default = NewLogger("app")

	// Enabled controls whether debug mode is active globally
	enabled   bool
	enabledMu sync.RWMutex
)

// NewLogger creates a new logger for a component
func NewLogger(component string) *Logger {
	return &Logger{
		level:         LevelInfo,
		output:        os.Stderr,
		enableStack:   false,
		stackDepth:    10,
		component:     component,
		showTimestamp: true,
		showCaller:    true,
	}
}

// Enable enables debug mode globally
func Enable() {
	enabledMu.Lock()
	enabled = true
	enabledMu.Unlock()
}

// Disable disables debug mode globally
func Disable() {
	enabledMu.Lock()
	enabled = false
	enabledMu.Unlock()
}

// IsEnabled returns whether debug mode is enabled
func IsEnabled() bool {
	enabledMu.RLock()
	defer enabledMu.RUnlock()
	return enabled
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	l.output = w
	l.mu.Unlock()
}

// SetLogFile sets output to a file
func (l *Logger) SetLogFile(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Close previous file if any
	if l.file != nil {
		l.file.Close()
	}

	l.file = f
	l.output = io.MultiWriter(os.Stderr, f)
	return nil
}

// EnableStackTraces enables full stack traces on errors
func (l *Logger) EnableStackTraces(depth int) {
	l.mu.Lock()
	l.enableStack = true
	if depth > 0 {
		l.stackDepth = depth
	}
	l.mu.Unlock()
}

// DisableStackTraces disables stack traces
func (l *Logger) DisableStackTraces() {
	l.mu.Lock()
	l.enableStack = false
	l.mu.Unlock()
}

// Close closes the log file if open
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if level > l.level {
		return
	}

	var sb strings.Builder

	// Timestamp
	if l.showTimestamp {
		sb.WriteString(time.Now().Format("2006-01-02 15:04:05.000"))
		sb.WriteString(" ")
	}

	// Level
	sb.WriteString("[")
	sb.WriteString(level.String())
	sb.WriteString("] ")

	// Component
	if l.component != "" {
		sb.WriteString("[")
		sb.WriteString(l.component)
		sb.WriteString("] ")
	}

	// Caller info
	if l.showCaller {
		_, file, line, ok := runtime.Caller(3)
		if ok {
			// Shorten path to just filename
			file = filepath.Base(file)
			sb.WriteString(fmt.Sprintf("%s:%d ", file, line))
		}
	}

	// Message
	if len(args) > 0 {
		sb.WriteString(fmt.Sprintf(format, args...))
	} else {
		sb.WriteString(format)
	}
	sb.WriteString("\n")

	// Stack trace for errors
	if l.enableStack && level <= LevelError {
		sb.WriteString(l.getStackTrace())
	}

	fmt.Fprint(l.output, sb.String())
}

func (l *Logger) getStackTrace() string {
	var sb strings.Builder
	sb.WriteString("Stack trace:\n")

	pcs := make([]uintptr, l.stackDepth)
	n := runtime.Callers(4, pcs) // Skip runtime.Callers, getStackTrace, log, and the public method
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		// Skip runtime internals
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}
		sb.WriteString(fmt.Sprintf("  %s\n    %s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	return sb.String()
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Trace logs a trace message
func (l *Logger) Trace(format string, args ...interface{}) {
	l.log(LevelTrace, format, args...)
}

// ErrorErr logs an error with the error object
func (l *Logger) ErrorErr(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.log(LevelError, "%s: %v", msg, err)
}

// WrapError wraps an error with stack trace info
func WrapError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)

	// Get caller info
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		return fmt.Errorf("%s:%d %s: %w", file, line, msg, err)
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// StackTrace returns the current stack trace as a string
func StackTrace(depth int) string {
	if depth <= 0 {
		depth = 10
	}

	var sb strings.Builder
	pcs := make([]uintptr, depth)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}
		sb.WriteString(fmt.Sprintf("  %s\n    %s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	return sb.String()
}

// Recover recovers from a panic and logs it with stack trace
func Recover(l *Logger) {
	if r := recover(); r != nil {
		l.Error("PANIC RECOVERED: %v\n%s", r, StackTrace(20))
	}
}

// DebugConfig holds debug configuration (matches storage.Config debug fields)
type DebugConfig struct {
	Enabled     bool
	Level       Level
	StackTraces bool
	LogToFile   bool
	LogFile     string
	DataDir     string // For default log file location
}

// SetupFromConfig configures the default logger from a DebugConfig
func SetupFromConfig(cfg *DebugConfig) {
	if cfg == nil {
		return
	}

	if cfg.Enabled {
		Enable()
		Default.SetLevel(cfg.Level)

		if cfg.StackTraces {
			Default.EnableStackTraces(15)
		}

		if cfg.LogToFile {
			logFile := cfg.LogFile
			if logFile == "" && cfg.DataDir != "" {
				logFile = filepath.Join(cfg.DataDir, "debug.log")
			}
			if logFile != "" {
				if err := Default.SetLogFile(logFile); err != nil {
					log.Printf("Failed to open log file: %v", err)
				}
			}
		}

		Default.Info("Debug mode enabled (level: %s, stack traces: %v)", cfg.Level, cfg.StackTraces)
	}
}

// SetupDefaultLogger configures the default logger based on environment
func SetupDefaultLogger() {
	// Check environment variables
	if os.Getenv("DISCORD_DEBUG") == "1" || os.Getenv("DEBUG") == "1" {
		Enable()
		Default.SetLevel(LevelDebug)
		Default.EnableStackTraces(15)
		log.Println("Debug mode enabled via environment variable")
	}

	if os.Getenv("DISCORD_TRACE") == "1" {
		Enable()
		Default.SetLevel(LevelTrace)
		Default.EnableStackTraces(20)
		log.Println("Trace mode enabled via environment variable")
	}

	// Log file
	if logFile := os.Getenv("DISCORD_LOG_FILE"); logFile != "" {
		if err := Default.SetLogFile(logFile); err != nil {
			log.Printf("Failed to open log file: %v", err)
		}
	}
}

// Package-level convenience functions using Default logger

func Error(format string, args ...interface{}) {
	Default.Error(format, args...)
}

func Warn(format string, args ...interface{}) {
	Default.Warn(format, args...)
}

func Info(format string, args ...interface{}) {
	Default.Info(format, args...)
}

func Debug(format string, args ...interface{}) {
	Default.Debug(format, args...)
}

func Trace(format string, args ...interface{}) {
	Default.Trace(format, args...)
}

func ErrorErr(err error, format string, args ...interface{}) {
	Default.ErrorErr(err, format, args...)
}
