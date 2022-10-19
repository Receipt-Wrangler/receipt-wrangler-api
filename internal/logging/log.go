package logging

import (
	"errors"
	"log"
	"os"
)

var logger *log.Logger

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

	// defer logFile.Close()

	// Flags are for date time, file name, and line number
	logger = log.New(logFile, "App", log.Lshortfile|log.LstdFlags)

	logger.Println("create")
	return nil
}

func GetLogger() *log.Logger {
	return logger
}
