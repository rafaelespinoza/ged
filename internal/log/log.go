// Package log is common structured logging utils.
package log

import (
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
)

var (
	ValidLoggingLevels  = []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	ValidLoggingFormats = []string{"JSON", "TEXT"}
)

func NewHandler(w io.Writer, loggingOff bool, logLevel, logFormat string) (slog.Handler, error) {
	if loggingOff {
		return slog.DiscardHandler, nil
	}

	var lvl slog.Level
	levels := make([]string, len(ValidLoggingLevels))
	for i, validLevel := range ValidLoggingLevels {
		levels[i] = validLevel.String()
	}
	if ind := slices.Index(levels, strings.ToUpper(strings.TrimSpace(logLevel))); ind >= 0 {
		lvl = ValidLoggingLevels[ind]
	} else {
		return nil, fmt.Errorf("invalid log level %q; should be one of %q", logLevel, ValidLoggingLevels)
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
		return nil, fmt.Errorf("invalid log format, should be one of %q", ValidLoggingFormats)
	}

	return handler, nil
}
