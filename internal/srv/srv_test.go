package srv

import (
	"fmt"
	"os"
	"testing"

	"github.com/rafaelespinoza/ged/internal/testutil"
)

func TestMain(m *testing.M) {
	reset, err := testutil.SetDefaultsFromCLI(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	defer reset()

	exitCode := m.Run()
	os.Exit(exitCode)
}
