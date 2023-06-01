package simpleutils

import "strconv"

func UintToString(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
