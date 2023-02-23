package utils

import "strconv"

func UintToString(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

func StringToUint64(v string) (uint64, error) {
	result, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, err
	}

	return result, nil
}
