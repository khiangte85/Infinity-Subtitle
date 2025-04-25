package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	logger      *log.Logger
	logFile     *os.File
	mu          sync.Mutex
	currentDate string
}

var (
	instance *Logger
	once     sync.Once
)

// GetLogger returns a singleton instance of Logger
func GetLogger() (*Logger, error) {
	var err error
	once.Do(func() {
		instance, err = newLogger()
	})
	return instance, err
}

func newLogger() (*Logger, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	currentDate := time.Now().Format("2006-01-02")
	logFileName := filepath.Join(logDir, currentDate+".log")
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		logger:      log.New(file, "", log.LstdFlags|log.Lmicroseconds),
		logFile:     file,
		currentDate: currentDate,
	}, nil
}

// checkAndRotate checks if we need to rotate the log file and does so if necessary
func (l *Logger) checkAndRotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	currentDate := time.Now().Format("2006-01-02")
	if currentDate == l.currentDate {
		return nil // No need to rotate
	}

	// Close the current file
	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			return fmt.Errorf("failed to close current log file: %w", err)
		}
	}

	// Create new log file
	logDir := "logs"
	logFileName := filepath.Join(logDir, currentDate+".log")
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	// Update logger with new file
	l.logger = log.New(file, "", log.LstdFlags|log.Lmicroseconds)
	l.logFile = file
	l.currentDate = currentDate

	return nil
}

// Close should be called when the logger is no longer needed
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// logWithRotation is a helper function that ensures log rotation is checked before logging
func (l *Logger) logWithRotation(level string, format string, v ...interface{}) {
	if err := l.checkAndRotate(); err != nil {
		// If rotation fails, log to stderr
		log.Printf("ERROR: Failed to rotate log file: %v", err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf(level+": "+format, v...)
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.logWithRotation("INFO", format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.logWithRotation("ERROR", format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.logWithRotation("DEBUG", format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.logWithRotation("WARN", format, v...)
}
