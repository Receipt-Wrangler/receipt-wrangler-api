package utils

import "regexp"

func GetTriggerRegex() regexp.Regexp {
	regex := regexp.MustCompile(`@\w+`)
	return *regex
}
