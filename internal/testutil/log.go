// Package testutil is test utility helpers.
package testutil

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/rafaelespinoza/ged/internal/log"
	"github.com/rafaelespinoza/logg"
)

// SetDefaultsFromCLI scans args for "-args", parses only the remaining flags,
// and then sets the global defaults in the log/slog package. It returns a
// reset function that should be called upon conclusion of the caller's tests,
// and any error from parsing the command line flags. The input, args, would be
// os.Args[1:].
// This func is intended to be used in [TestMain] while invoking `go test` with
// particular command line flags. See the example for more.
//
// [TestMain]: https://pkg.go.dev/testing#hdr-Main.
func SetDefaultsFromCLI(args []string) (reset func(), err error) {
	var logger LogSetup
	if err := logger.parseFlags(args); err != nil {
		return nil, err
	}
	return logger.setDefaults(), nil
}

// SetScopedDefaults replaces the logger for per-test isolation using t.Cleanup.
func SetScopedDefaults(tb testing.TB, s LogSetup) {
	tb.Helper()

	nextHandler := s.Handler
	if nextHandler == nil {
		nextHandler = slog.NewTextHandler(
			tb.Output(),
			&slog.HandlerOptions{Level: slog.LevelInfo},
		)
	}
	reset := s.setDefaults()
	tb.Cleanup(reset)
}

// LogSetup is a set of named configurations for a test.
type LogSetup struct {
	Handler slog.Handler

	// these fields capture command line values

	format string
	level  string
	quiet  bool
}

func (s *LogSetup) parseFlags(args []string) error {
	// Locate "-args" in the slice
	argsIndex := slices.Index(args, "-args")
	if argsIndex == -1 {
		// "-args" was not passed; keep default settings
		return nil
	}

	customArgs := args[argsIndex+1:]
	fs := flag.NewFlagSet("ged-testutil", flag.ContinueOnError)

	fs.BoolVar(&s.quiet, "q", false, "if true, then all logging is effectively off")
	fs.StringVar(&s.level, "loglevel", log.ValidLoggingLevels[len(log.ValidLoggingLevels)-1].String(), fmt.Sprintf("minimum severity for which to log events, should be one of %q", log.ValidLoggingLevels))
	fs.StringVar(&s.format, "logformat", log.ValidLoggingFormats[len(log.ValidLoggingFormats)-1], fmt.Sprintf("output format for logs, should be one of %q", log.ValidLoggingFormats))
	if err := fs.Parse(customArgs); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if s.quiet {
		s.Handler = slog.DiscardHandler
		return nil
	}

	handler, err := log.NewHandler(os.Stderr, false, s.level, s.format)
	if err != nil {
		return fmt.Errorf("making new log handler: %w", err)
	}
	s.Handler = handler
	return nil
}

func (s *LogSetup) setDefaults() (reset func()) {
	prevDefault := slog.Default()
	// Use the same pattern from the cmd package.
	logg.SetDefaults(s.Handler, nil)
	return func() { slog.SetDefault(prevDefault) }
}
