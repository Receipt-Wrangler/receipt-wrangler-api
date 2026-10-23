package emailclient

import (
	"fmt"
	"os"
	"testing"

	"receipt-wrangler/api/internal/logging"
)

// TestMain initializes the logger, which writes to a relative logs/ directory, from
// inside a scratch directory so the run leaves nothing behind in the package.
// This package needs neither the database nor Redis.
func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return 1
	}

	scratch, err := os.MkdirTemp("", "emailclient-test-")
	if err != nil {
		fmt.Println(err)
		return 1
	}
	defer os.RemoveAll(scratch)

	if err := os.Chdir(scratch); err != nil {
		fmt.Println(err)
		return 1
	}
	defer os.Chdir(workingDirectory)

	if err := logging.InitLog(); err != nil {
		fmt.Println(err)
		return 1
	}

	return m.Run()
}
