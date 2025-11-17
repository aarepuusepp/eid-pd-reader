package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile   *os.File
	logMutex  sync.Mutex
	logWriter io.Writer
	enabled   bool
)

// Init initializes the logger with a log file path
func Init(logPath string) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	// If no path specified, use default location
	if logPath == "" {
		// Use %LOCALAPPDATA%\eid-pd-reader\eid-pd-reader.log on Windows
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			appData = os.Getenv("APPDATA")
		}
		if appData == "" {
			// Fallback to temp directory
			appData = os.TempDir()
		}
		
		logDir := filepath.Join(appData, "eid-pd-reader")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		
		logPath = filepath.Join(logDir, "eid-pd-reader.log")
	}

	// Open log file (append mode)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Close old log file if exists
	if logFile != nil {
		logFile.Close()
	}

	logFile = f
	logWriter = f
	enabled = true

	// Log initialization (don't use Log() to avoid deadlock)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] [INFO] Logger initialized | logPath=%s pid=%d\n", timestamp, logPath, os.Getpid())
	f.Write([]byte(logLine))

	return nil
}

// Close closes the log file
func Close() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		logFile.Close()
		logFile = nil
		logWriter = nil
	}
	enabled = false
}

// Log writes a log entry
func Log(level, message string, data map[string]interface{}) {
	if !enabled {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	// Build log line
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
	
	// Add data fields
	if len(data) > 0 {
		logLine += " |"
		for k, v := range data {
			logLine += fmt.Sprintf(" %s=%v", k, v)
		}
	}
	
	logLine += "\n"
	
	// Write to log file
	if logWriter != nil {
		logWriter.Write([]byte(logLine))
	}
}

// Info logs an info message
func Info(message string, data map[string]interface{}) {
	Log("INFO", message, data)
}

// Error logs an error message
func Error(message string, data map[string]interface{}) {
	Log("ERROR", message, data)
}

// Warn logs a warning message
func Warn(message string, data map[string]interface{}) {
	Log("WARN", message, data)
}

// Debug logs a debug message
func Debug(message string, data map[string]interface{}) {
	Log("DEBUG", message, data)
}

// LogCardRead logs a successful card read operation
func LogCardRead(personalCode, surname, givenName, documentNr, source string) {
	Info("Card read successful", map[string]interface{}{
		"personalCode": personalCode,
		"surname":      surname,
		"givenName":    givenName,
		"documentNr":   documentNr,
		"source":       source,
	})
}

// LogCardReadError logs a failed card read operation
func LogCardReadError(errorMsg string) {
	Error("Card read failed", map[string]interface{}{
		"error": errorMsg,
	})
}

// LogNativeMessage logs a native messaging request/response
func LogNativeMessage(direction, command string, success bool, errorMsg string) {
	level := "INFO"
	if !success {
		level = "ERROR"
	}
	
	data := map[string]interface{}{
		"direction": direction,
		"command":   command,
		"success":   success,
	}
	
	if errorMsg != "" {
		data["error"] = errorMsg
	}
	
	Log(level, "Native message", data)
}

// GetLogPath returns the current log file path
func GetLogPath() string {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	if logFile != nil {
		return logFile.Name()
	}
	return ""
}

// IsEnabled returns whether logging is enabled
func IsEnabled() bool {
	return enabled
}

