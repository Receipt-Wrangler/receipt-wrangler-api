package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var logger *log.Logger
var stdLogger *log.Logger

func InitLog() error {
	logPath := "logs/app.log"
	logDir := "logs"
	if _, err := os.Stat(logDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logDir, os.ModePerm)
		if err != nil {
			log.Print(err.Error())
			return err
		}
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	logFlags := log.Lshortfile | log.LstdFlags

	// Flags are for date time, file name, and line number
	logger = log.New(logFile, "App", logFlags)
	stdLogger = log.New(os.Stdout, "App", logFlags)

	return nil
}

func GetLogger() *log.Logger {
	return logger
}

func LogStd(level LogLevel, v ...any) {
	_, file, line, _ := runtime.Caller(1)
	lineInfo := fmt.Sprintf("[%s:%d]", filepath.Base(file), line)
	levelString := fmt.Sprintf("%s: ", level)
	v = append([]any{lineInfo, levelString}, v...)

	if level == LOG_LEVEL_FATAL {
		logger.Println(v...)
		stdLogger.Println(v...)
		os.Exit(1)
	}
	if level == LOG_LEVEL_ERROR {
		logger.Println(v...)
		stdLogger.Println(v...)
	}
	if level == LOG_LEVEL_INFO {
		logger.Println(v...)
		stdLogger.Println(v...)
	}
}
