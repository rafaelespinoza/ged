package testutil_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/rafaelespinoza/ged/internal/testutil"
)

// This is an example of setting the global logging defaults via the command
// line. what you might put into [TestMain]. It demonstrates usage of the -args
// flag of go test. For more info see "go help test" and "go help testflag".
//
//	$ go test -v -count=1 ./internal/gedcom -- -args -loglevel=DEBUG
//
// [TestMain]: https://pkg.go.dev/testing#hdr-Main.
func ExampleSetDefaultsFromCLI() {
	// This would be made available when you define a TestMain in your test.
	var m *testing.M

	// Simulate command-line arguments: go test ... -args -loglevel=DEBUG
	osArgs := []string{"go", "test", "-v", "-count", "1", "./pkg/path", "--", "-args", "-loglevel", "DEBUG"}

	reset, err := testutil.SetDefaultsFromCLI(osArgs[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	defer reset()

	slog.Debug("This debug message is captured")

	exitCode := m.Run()
	os.Exit(exitCode)
}

// Demonstrate how to directly override the log level for a specific test
// without needing to parse command line flags.
func ExampleSetScopedDefaults() {
	// This would be made available when you define a regular Test* function.
	var t *testing.T

	handler := slog.NewTextHandler(
		t.Output(),
		&slog.HandlerOptions{
			Level: slog.LevelWarn,
		},
	)
	testutil.SetScopedDefaults(t, testutil.LogSetup{Handler: handler})

	// INFO is suppressed because level is set to WARN
	slog.Info("suppressed info message")
	slog.Warn("visible warning message")
}
