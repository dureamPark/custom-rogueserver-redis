package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel defines the severity of a log message.
type LogLevel int

const (
	INFO LogLevel = iota
	WARN
	ERROR
	DEBUG
)

var (
	logLevelNames = []string{"INFO", "WARN", "ERROR", "DEBUG"}
	mu            sync.Mutex
	logger        *log.Logger
	currentLevel  LogLevel = INFO // Default log level
)

func init() {
	// Initialize the standard logger to write to os.Stdout
	logger = log.New(os.Stdout, "", 0) // No default flags, we'll format manually
}

// SetLogLevel sets the minimum log level to be displayed.
func SetLogLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = level
}

// formatMessage formats the log message with timestamp, file/line, and level.
func formatMessage(level LogLevel, format string, args ...interface{}) string {
	now := time.Now().Format("2006-01-02 15:04:05.000")
	_, file, line, ok := runtime.Caller(2) // Caller(2) to get the original caller's file/line
	if !ok {
		file = "???"
		line = 0
	} else {
		// Shorten file path
		parts := strings.Split(file, "/")
		file = parts[len(parts)-1]
	}

	levelStr := logLevelNames[level]
	message := fmt.Sprintf(format, args...)

	return fmt.Sprintf("[%s] %s %s:%d - %s", now, levelStr, file, line, message)
}

// Info logs an informational message.
func Info(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if currentLevel <= INFO {
		logger.Println(formatMessage(INFO, format, args...))
	}
}

// Warn logs a warning message.
func Warn(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if currentLevel <= WARN {
		logger.Println(formatMessage(WARN, format, args...))
	}
}

// Error logs an error message.
func Error(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if currentLevel <= ERROR {
		logger.Println(formatMessage(ERROR, format, args...))
	}
}

// Debug logs a debug message.
func Debug(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if currentLevel <= DEBUG {
		logger.Println(formatMessage(DEBUG, format, args...))
	}
}
