package utils

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (code int, err error) {

	defer teardown()

	return m.Run(), nil
}

func teardown() {
}
