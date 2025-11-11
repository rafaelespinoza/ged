package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/rafaelespinoza/alf"

	"github.com/rafaelespinoza/logg"
)

var (
	args struct {
		loggingOff bool
		logLevel   string
		logFormat  string
	}
)

const mainName = "ged"

// Root abstracts a top-level command from package main.
type Root interface {
	// Run is the entry point. It should be called with os.Args[1:].
	Run(ctx context.Context, args []string) error
}

// New constructs a top-level command with subcommands.
func New() Root {
	del := &alf.Delegator{
		Description: "main command for " + mainName,
		Subs: map[string]alf.Directive{
			"draw":         makeDraw("draw"),
			"parse":        makeParse("parse"),
			"explore-data": makeExploreData("explore-data"),
		},
	}

	rootFlags := newFlagSet(mainName)
	rootFlags.BoolVar(&args.loggingOff, "q", false, "if true, then all logging is effectively off")
	rootFlags.StringVar(&args.logLevel, "loglevel", validLoggingLevels[len(validLoggingLevels)-1].String(), fmt.Sprintf("minimum severity for which to log events, should be one of %q", validLoggingLevels))
	rootFlags.StringVar(&args.logFormat, "logformat", validLoggingFormats[len(validLoggingFormats)-1], fmt.Sprintf("output format for logs, should be one of %q", validLoggingFormats))

	rootFlags.Usage = func() {
		fmt.Fprintf(rootFlags.Output(), `%s

Description:

	%s processes genealogical data in GEDCOM format.
	GEDCOM is a defacto standard format, read more at https://gedcom.io

	There are top-level flags, defined here, which may apply to any subcommand.
	Logging messages may be useful for extra runtime introspection as the
	program traverses the input data or is calculating something. Messages are
	structured data, written to STDERR. Messages have an associated "Level", to
	describe the severity of an event.

	Each subcommand may have its own flags, which would be defined there.

	The following flags are top-level and should go before the subcommand.
`,
			initUsageLine("subcommand"), mainName)
		printFlagDefaults(rootFlags)

		fmt.Fprintf(
			rootFlags.Output(), `
Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v

Examples:

	%s [%s-flags] <subcommand> -h
	%s -loglevel INFO <subcommand> [subcommand-flags]
	%s -logformat json <subcommand> [subcommand-flags]
	%s -logformat text -loglevel WARN <subcommand> [subcommand-flags]
`,
			strings.Join(del.DescribeSubcommands(), "\n\t"), mainName, mainName, mainName, mainName, mainName)
	}

	del.Flags = rootFlags

	return &alf.Root{
		Delegator: del,
		PrePerform: func(_ context.Context) error {
			handler, err := newLogHandler(os.Stderr, args.loggingOff, args.logLevel, args.logFormat)
			if err != nil {
				return err
			}

			logg.SetDefaults(handler, nil)
			return nil
		},
	}
}

var (
	validLoggingLevels  = []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	validLoggingFormats = []string{"JSON", "TEXT"}
)

func newLogHandler(w io.Writer, loggingOff bool, logLevel, logFormat string) (slog.Handler, error) {
	if loggingOff {
		return slog.NewTextHandler(io.Discard, nil), nil
	}

	var lvl slog.Level
	levels := make([]string, len(validLoggingLevels))
	for i, validLevel := range validLoggingLevels {
		levels[i] = validLevel.String()
	}
	if ind := slices.Index(levels, strings.ToUpper(strings.TrimSpace(logLevel))); ind >= 0 {
		lvl = validLoggingLevels[ind]
	} else {
		return nil, fmt.Errorf("invalid log level %q; should be one of %q", logLevel, validLoggingLevels)
	}

	opts := slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
	var handler slog.Handler

	switch strings.ToUpper(strings.TrimSpace(logFormat)) {
	case "JSON":
		handler = slog.NewJSONHandler(w, &opts)
	case "TEXT":
		handler = slog.NewTextHandler(w, &opts)
	default:
		return nil, fmt.Errorf("invalid log format, should be one of %q", validLoggingFormats)
	}

	return handler, nil
}

func newFlagSet(name string) (out *flag.FlagSet) {
	out = flag.NewFlagSet(name, flag.ExitOnError)
	out.SetOutput(os.Stdout)
	return
}

// printFlagDefaults calls PrintDefaults on f. It helps make help message
// formatting more consistent.
func printFlagDefaults(f *flag.FlagSet) {
	fmt.Fprintf(f.Output(), "\nFlags for %s:\n\n", f.Name())
	f.PrintDefaults()
}

func readJSON(in io.Reader, out any) error    { return json.NewDecoder(in).Decode(out) }
func writeJSON(out io.Writer, data any) error { return json.NewEncoder(out).Encode(data) }

func initUsageLine(subcmd string) string {
	return fmt.Sprintf("Usage: %s [%s-flags] %s [%s-flags]", mainName, mainName, subcmd, subcmd)
}
