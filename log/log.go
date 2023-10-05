package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

// LogLevel represents the log level.
type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

var (
	logFile         *os.File
	dir             string   = "logs"
	pattern         string   = "2006-01-02 15.04.05"
	logLevel        LogLevel = TRACE
	logLevelStrings          = []string{
		"TRACE",
		"DEBUG",
		"INFO",
		"WARNING",
		"ERROR",
		"FATAL",
	}
)

func Initialize(dirx, patternx, level string) {
	dir = dirx
	pattern = patternx
	SetLogLevel(level)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	Rotate()
}

func (level LogLevel) String() string {
	return logLevelStrings[level]
}

// logMessage is a helper function to log a message with the given log level.
func logMessage(level LogLevel, format string, v ...any) {
	if level >= logLevel {
		_, file, line, _ := runtime.Caller(2)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		log.Printf(
			"[%s] %s (%s:%d)", level, fmt.Sprintf(format, v...), file, line,
		)
	}
}

// Trace logs a message with TRACE log level.
func Trace(format string, v ...any) {
	logMessage(TRACE, format, v...)
}

// Debug logs a message with DEBUG log level.
func Debug(format string, v ...any) {
	logMessage(DEBUG, format, v...)
}

// Info logs a message with INFO log level.
func Info(format string, v ...any) {
	logMessage(INFO, format, v...)
}

// Warning logs a message with WARNING log level.
func Warning(format string, v ...any) {
	logMessage(WARNING, format, v...)
}

// Error logs a message with ERROR log level.
func Error(format string, v ...any) {
	logMessage(ERROR, format, v...)
}

// Fatal logs a message with FATAL log level.
func Fatal(format string, v ...any) {
	logMessage(FATAL, format, v...)
}

// SetLogLevel sets the log level.
func SetLogLevel(level string) {
	for i, s := range logLevelStrings {
		if s == level {
			logLevel = LogLevel(i)
			return
		}
	}
}

/*
log.SetOutput and writing functions are using same mutex so SetOutput is
Thread safe.
*/
func Rotate() {
	fname := time.Now().Format(dir + "/" + pattern + ".log")
	w, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, w))

	logFile = w
	Trace("Log rotated: %v", fname)
}
