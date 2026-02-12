package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var debugFilename = "debug-ripple.txt"

const errorPrefix = "Ripple had a fatal error %s"

const (
	reset    = "\033[0m"
	black    = "\033[30m"
	red      = "\033[31m"
	green    = "\033[32m"
	yellow   = "\033[33m"
	blue     = "\033[34m"
	purple   = "\033[35m"
	cyan     = "\033[36m"
	white    = "\033[37m"
	darkGray = "\033[90m"
)

var logger = new()

type Logger struct {
	debugEnabled bool
}

func new() *Logger {
	f, err := os.OpenFile(debugFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.WriteString("\n")
	f.WriteString(addTimestamp("--- New Logging Entries --- \n"))

	return &Logger{
		debugEnabled: false,
	}
}

func Init(debugMode bool) {
	logger.debugEnabled = debugMode
}

//todo: we should be able add the caller to the log

func (l *Logger) info(str string, params ...any) {
	msg := fmt.Sprintf(str, params...)
	log.Println(colorize(blue, msg))

	l.writeToFile(msg)
}

func (l *Logger) debug(str string, params ...any) {
	if !l.debugEnabled {
		return
	}

	msg := fmt.Sprintf(str, params...)
	log.Println(colorize(green, msg))

	l.writeToFile(msg)
}

func (l *Logger) error(str string, params ...any) {
	errMsg := fmt.Sprintf(errorPrefix, fmt.Sprintf(str, params...))
	log.Println(colorize(red, errMsg))

	l.writeToFile(errMsg)
}

func (l *Logger) writeToFile(entry string) {
	f, err := os.OpenFile(debugFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		log.Println(colorize(red, err.Error()))
	}

	defer f.Close()

	if _, err = fmt.Fprintf(f, "%s \n", addTimestamp(entry)); err != nil {
		log.Println(colorize(red, err.Error()))
	}
}

func addTimestamp(message string) string {
	return fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)
}

func colorize(color, message string) string {
	return fmt.Sprintf("%s%s%s", color, message, reset)
}

func Info(str string, params ...any) {
	logger.info(str, params...)
}

func Debug(str string, params ...any) {
	logger.debug(str, params...)
}

func Error(str string, params ...any) {
	logger.error(str, params...)
}
