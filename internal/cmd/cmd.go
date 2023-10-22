package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/rafaelespinoza/alf"

	"github.com/rafaelespinoza/ged/internal/log"
)

var (
	args struct {
		loggingOff bool
		logLevel   string
		logFormat  string
	}
	// bin is the name of the binary.
	bin = os.Args[0]
)

// Root abstracts a top-level command from package main.
type Root interface {
	// Run is the entry point. It should be called with os.Args[1:].
	Run(ctx context.Context, args []string) error
}

// New constructs a top-level command with subcommands.
func New() Root {
	del := &alf.Delegator{
		Description: "main command for " + bin,
		Subs: map[string]alf.Directive{
			"draw":   makeDraw("draw"),
			"parse":  makeParse("parse"),
			"relate": makeRelate("relate"),
		},
	}

	rootFlags := newFlagSet("ged")
	rootFlags.BoolVar(&args.loggingOff, "q", false, "if true, then all logging is effectively off")
	rootFlags.StringVar(&args.logLevel, "loglevel", validLoggingLevels[len(validLoggingLevels)-1].String(), fmt.Sprintf("minimum severity for which to log events, should be one of %q", validLoggingLevels))
	rootFlags.StringVar(&args.logFormat, "logformat", validLoggingFormats[len(validLoggingFormats)-1], fmt.Sprintf("output format for logs, should be one of %q", validLoggingFormats))
	rootFlags.Usage = func() {
		fmt.Fprintf(rootFlags.Output(), `Usage: %s

	The following flags should go before the command.
`,
			bin)
		printFlagDefaults(rootFlags)
		fmt.Fprintf(
			rootFlags.Output(), `
Commands:

	These will have their own set of flags. Put them after the command.

	%v

Examples:

	%s [ged-flags] <command> -h
	%s -loglevel INFO <command>
	%s -logformat json <command>
	%s -logformat text -loglevel WARN <command>
`,
			strings.Join(del.DescribeSubcommands(), "\n\t"), bin, bin, bin, bin)
	}

	del.Flags = rootFlags

	return &alf.Root{
		Delegator: del,
		PrePerform: func(_ context.Context) error {
			handler, err := newLogHandler(os.Stderr)
			if err != nil {
				return err
			}
			log.Init(handler)
			return nil
		},
	}
}

var (
	validLoggingLevels  = []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	validLoggingFormats = []string{"json", "text"}
)

func newLogHandler(w io.Writer) (slog.Handler, error) {
	if args.loggingOff {
		return nil, nil
	}

	lvl := slog.LevelDebug - 1 // sentinel value to help recognize invalid input
	for _, validLevel := range validLoggingLevels {
		if strings.ToUpper(args.logLevel) == validLevel.String() {
			lvl = validLevel
			break
		}
	}
	if lvl < slog.LevelDebug {
		return nil, fmt.Errorf("invalid log level %q; should be one of %q", args.logLevel, validLoggingLevels)
	}

	replaceAttrs := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}

	var h slog.Handler
	switch strings.ToLower(args.logFormat) {
	case "json":
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl, ReplaceAttr: replaceAttrs})
	case "text":
		h = slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl, ReplaceAttr: replaceAttrs})
	default:
		return nil, fmt.Errorf("invalid logformat, should be one of %q", validLoggingFormats)
	}

	return h, nil
}

func newFlagSet(name string) (out *flag.FlagSet) {
	out = flag.NewFlagSet(name, flag.ExitOnError)
	out.SetOutput(os.Stdout)
	return
}

// printFlagDefaults calls PrintDefaults on f. It helps make help message
// formatting more consistent.
func printFlagDefaults(f *flag.FlagSet) {
	fmt.Fprintf(f.Output(), "\n%s flags:\n\n", f.Name())
	f.PrintDefaults()
}

func readJSON(in io.Reader, out any) error    { return json.NewDecoder(in).Decode(out) }
func writeJSON(out io.Writer, data any) error { return json.NewEncoder(out).Encode(data) }
