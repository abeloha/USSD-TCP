package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogLevel int

const (
	INFO LogLevel = iota
	WARN
	ERROR
	DEBUG
)

type Logger struct {
	logFile   *os.File
	logPath   string
	logPrefix string
}

func New(logPath string) (*Logger, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		return nil, err
	}

	// Create log file for current date
	currentDate := time.Now().Format("2006-01-02")
	filename := filepath.Join(logPath, currentDate+".log")
	
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		logFile:   logFile,
		logPath:   logPath,
		logPrefix: "[USSDTCP]",
	}, nil
}

func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	levelPrefix := map[LogLevel]string{
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
		DEBUG: "DEBUG",
	}[level]

	logEntry := fmt.Sprintf("%s %s %s: %s\n", 
		time.Now().Format(time.RFC3339), 
		l.logPrefix, 
		levelPrefix, 
		fmt.Sprintf(format, v...),
	)

	// Write to file
	if _, err := l.logFile.WriteString(logEntry); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}

	// Also log to console
	log.Printf("%s %s: %s", l.logPrefix, levelPrefix, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WARN, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

func (l *Logger) Close() error {
	return l.logFile.Close()
}