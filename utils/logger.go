package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	*log.Logger
	level LogLevel
}

var (
	instance *Logger
	once     sync.Once
)

func getLogLevel() LogLevel {
	level := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// GetLogger returns a singleton instance of Logger
func GetLogger() *Logger {
	once.Do(func() {
		var err error
		instance, err = newLogger()
		if err != nil {
			// Fallback to basic stderr logging if file creation fails
			instance = &Logger{
				Logger: log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile),
				level:  getLogLevel(),
			}
		}
	})
	return instance
}

// newLogger is now private
func newLogger() (*Logger, error) {
	var writer io.Writer = os.Stderr

	// Try to create log file, fallback to stderr if fails
	logDir := "/var/log/email-printer"
	if os.Getuid() != 0 {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			logDir = filepath.Join(homeDir, ".local", "share", "email-printer", "logs")
		}
	}

	// Attempt to create log file, but don't fail if we can't
	if err := os.MkdirAll(logDir, 0755); err == nil {
		logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
		if f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); err == nil {
			writer = io.MultiWriter(f, os.Stderr)
		}
	}

	return &Logger{
		Logger: log.New(writer, "", log.Ldate|log.Ltime|log.Lshortfile),
		level:  getLogLevel(),
	}, nil
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.Printf("[DEBUG] "+format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.Printf("[INFO] "+format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.Printf("[WARN] "+format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.Printf("[ERROR] "+format, v...)
	}
}
