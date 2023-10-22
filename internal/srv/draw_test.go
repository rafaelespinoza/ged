package srv

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/rafaelespinoza/ged/internal/entity"
)

func TestDraw(t *testing.T) {
	tests := []struct {
		Name   string
		Params DrawParams
	}{
		{
			Name:   "no people, no unions",
			Params: DrawParams{},
		},
		{
			Name: "some people, no unions",
			Params: DrawParams{
				People: []*entity.Person{{}, {}},
			},
		},
		{
			Name: "no people, some unions",
			Params: DrawParams{
				Unions: []*entity.Union{{}, {}},
			},
		},
		{
			Name: "some people, some unions",
			Params: DrawParams{
				People: []*entity.Person{{}, {}},
				Unions: []*entity.Union{{}, {}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			sink := new(bytes.Buffer)
			test.Params.Out = sink

			err := Draw(context.Background(), test.Params)
			if err != nil {
				t.Fatal(err)
			}

			got, err := io.ReadAll(sink)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) < 1 {
				t.Fatal("does not seem like any output was written")
			}
		})
	}
}
