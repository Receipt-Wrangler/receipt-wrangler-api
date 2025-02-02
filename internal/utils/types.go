package utils

import (
	"strconv"
	"strings"
)

func UintToString(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

func StringToUint(v string) (uint, error) {
	vTrimmed := strings.Trim(v, " ")
	result, err := strconv.ParseUint(vTrimmed, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(result), nil
}

func StringToUint64(v string) (uint64, error) {
	result, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func StringToInt(v string) (int, error) {
	result, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return result, nil
}
