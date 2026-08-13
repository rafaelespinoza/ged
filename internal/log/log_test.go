package log_test

import (
	"testing"

	"github.com/rafaelespinoza/ged/internal/log"
)

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name       string
		off        bool
		level      string
		format     string
		expHandler bool
		expErr     bool
	}{
		{name: "logging off", off: true},
		{name: "invalid level", level: "bad", format: "JSON", expErr: true},
		{name: "invalid format", level: "INFO", format: "bad", expErr: true},
		{name: "ok JSON", level: "INFO", format: "JSON"},
		{name: "ok json", level: "INFO", format: "json"},
		{name: "ok TEXT", level: "INFO", format: "TEXT"},
		{name: "ok text", level: "INFO", format: "text"},
		{name: "ok tExT", level: "INFO", format: "tExT"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, err := log.NewHandler(t.Output(), test.off, test.level, test.format)

			if test.expErr && err == nil {
				t.Fatal("expected an error")
			} else if !test.expErr && err != nil {
				t.Fatalf("unexpected error %v", err)
			} else if test.expErr && err != nil {
				return // ok
			}

			if handler == nil {
				t.Fatal("expected non-empty handler")
			}
		})
	}
}
