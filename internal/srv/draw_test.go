package srv

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/rafaelespinoza/ged/internal/entity"
)

func TestDraw(t *testing.T) {
	validDefaultDirection := ValidFlowchartDirections[0]

	t.Run("outputs something", func(t *testing.T) {
		tests := []struct {
			Name   string
			Params DrawParams
		}{
			{
				Name:   "no people, no unions",
				Params: DrawParams{Direction: validDefaultDirection},
			},
			{
				Name: "some people, no unions",
				Params: DrawParams{
					Direction: validDefaultDirection,
					People:    []*entity.Person{{}, {}},
				},
			},
			{
				Name: "no people, some unions",
				Params: DrawParams{
					Direction: validDefaultDirection,
					Unions:    []*entity.Union{{}, {}},
				},
			},
			{
				Name: "some people, some unions",
				Params: DrawParams{
					Direction: validDefaultDirection,
					People:    []*entity.Person{{}, {}},
					Unions:    []*entity.Union{{}, {}},
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
	})

	t.Run("DisplayID", func(t *testing.T) {
		for _, displayID := range []bool{true, false} {
			sink := new(strings.Builder)

			people := []*entity.Person{
				{ID: "@IFoo@", Name: entity.PersonalName{Forename: "Foxtrot", Surname: "Foo"}},
				{ID: "@IBar@", Name: entity.PersonalName{Forename: "Bravo", Surname: "Bar"}},
			}
			err := Draw(context.Background(), DrawParams{
				Direction: validDefaultDirection,
				Out:       sink,
				DisplayID: displayID,
				People:    people,
			})
			if err != nil {
				t.Fatal(err)
			}

			got := sink.String()
			if len(got) < 1 {
				t.Fatal("does not seem like any output was written")
			}

			for _, id := range []string{"@IFoo@", "@IBar@"} {
				if !displayID {
					if strings.Contains(got, id) {
						t.Errorf("did not expect to find id %q in output", id)
					}
				} else {
					if !strings.Contains(got, id) {
						t.Errorf("expected to find id %q in output", id)
					}
				}
			}
		}
	})

	t.Run("Direction", func(t *testing.T) {
		for _, direction := range ValidFlowchartDirections {
			t.Run(direction, func(t *testing.T) {
				sink := new(bytes.Buffer)

				err := Draw(context.Background(), DrawParams{Direction: direction, Out: sink})
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

				expectedDirection := "flowchart " + direction + "\n"
				if !bytes.Contains(got, []byte(expectedDirection)) {
					t.Errorf("expected flowchart to contain %q", expectedDirection)
					t.Logf("for reference, here is flowchart\n%s", got)
				}
			})
		}

		t.Run("error", func(t *testing.T) {
			sink := new(bytes.Buffer)

			err := Draw(context.Background(), DrawParams{Direction: "invalid", Out: sink})
			if err == nil {
				t.Error("expected an error but got nil")
			} else if !strings.Contains(err.Error(), "invalid Direction") {
				t.Errorf("expected error message (%q) to contain %q", err.Error(), "invalid Direction")
			}

			got, err := io.ReadAll(sink)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) > 0 {
				t.Fatal("unexpected output")
			}
		})
	})
}
