package logging

import (
	"database/sql/driver"
	"errors"
)

type LogLevel string

const (
	LOG_LEVEL_INFO  LogLevel = "INFO"
	LOG_LEVEL_ERROR LogLevel = "ERROR"
	LOG_LEVEL_FATAL LogLevel = "FATAL"
)

func (logLevel *LogLevel) Scan(value string) error {
	*logLevel = LogLevel(value)
	return nil
}

func (logLevel LogLevel) Value() (driver.Value, error) {
	if logLevel != LOG_LEVEL_INFO &&
		logLevel != LOG_LEVEL_ERROR &&
		logLevel != LOG_LEVEL_FATAL {
		return nil, errors.New("invalid LogLevel")
	}
	return string(logLevel), nil
}
