package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"

	"github.com/rafaelespinoza/reltree/internal/log"
	"github.com/rafaelespinoza/reltree/internal/srv"
)

func init() {
	replaceAttrs := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}

	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo, ReplaceAttr: replaceAttrs})
	log.Init(h)
}

func main() {
	people, unions, err := srv.ParseGedcom(context.Background(), os.Stdin)
	if err != nil {
		panic(err)
	}
	writeJSON(os.Stdout, map[string]any{"people": people, "unions": unions})
}

func writeJSON(out io.Writer, data any) error { return json.NewEncoder(out).Encode(data) }
