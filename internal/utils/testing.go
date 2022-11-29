package utils

import (
	"testing"
)

func printTestError(t *testing.T, actual any, expected any) {
	t.Errorf("Expected %s, but got %s", expected, actual)
}
