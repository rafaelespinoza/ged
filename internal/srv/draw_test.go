package srv

import (
	"bytes"
	"context"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/rafaelespinoza/ged/internal/entity"
)

func TestMakeMermaidFlowchart(t *testing.T) {
	validDefaultDirection := ValidFlowchartDirections[0]

	t.Run("outputs something", func(t *testing.T) {
		tests := []struct {
			Name   string
			Params MermaidFlowchartParams
		}{
			{
				Name:   "no people, no unions",
				Params: MermaidFlowchartParams{Direction: validDefaultDirection},
			},
			{
				Name: "some people, no unions",
				Params: MermaidFlowchartParams{
					Direction: validDefaultDirection,
					People:    []*entity.Person{{}, {}},
				},
			},
			{
				Name: "no people, some unions",
				Params: MermaidFlowchartParams{
					Direction: validDefaultDirection,
					Unions:    []*entity.Union{{}, {}},
				},
			},
			{
				Name: "some people, some unions",
				Params: MermaidFlowchartParams{
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

				err := MakeMermaidFlowchart(context.Background(), test.Params)
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
			err := MakeMermaidFlowchart(context.Background(), MermaidFlowchartParams{
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

				err := MakeMermaidFlowchart(context.Background(), MermaidFlowchartParams{Direction: direction, Out: sink})
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

			err := MakeMermaidFlowchart(context.Background(), MermaidFlowchartParams{Direction: "invalid", Out: sink})
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

func TestMermaidRenderer(t *testing.T) {
	const input = `flowchart LR

%% define people

	BART_SIMPSON("Bart Simpson")
	HOMER_SIMPSON("Homer Simpson")
	LISA_SIMPSON("Lisa Simpson")
	MAGGIE_SIMPSON("Maggie Simpson")
	MARGE_SIMPSON("Marge Simpson")

%% define unions

	%% "Homer Simpson" and "Marge Simpson"
	F0000>"
H. Simpson
+
M. Simpson
"]

	HOMER_SIMPSON-...->F0000
	MARGE_SIMPSON-...->F0000

	F0000 =====> BART_SIMPSON
	F0000 =====> MAGGIE_SIMPSON
	F0000 =====> LISA_SIMPSON`

	t.Run("SVG", func(t *testing.T) {
		testSVG(t, context.Background(), strings.NewReader(input))
	})

	t.Run("PNG", func(t *testing.T) {
		testPNG(t, context.Background(), strings.NewReader(input), 5.0)
	})
}

func TestDraw(t *testing.T) {
	gedcomToMermaid := func(t *testing.T, ctx context.Context, pathToFile string) io.Reader {
		file, err := os.Open(filepath.Clean(pathToFile))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = file.Close() }()

		people, unions, err := ParseGedcom(ctx, file)
		if err != nil {
			t.Fatal(err)
		}

		buf := new(bytes.Buffer)
		params := MermaidFlowchartParams{
			Out:       buf,
			Direction: "LR",
			DisplayID: true,
			People:    people,
			Unions:    unions,
		}
		if err = MakeMermaidFlowchart(ctx, params); err != nil {
			t.Fatal(err)
		}

		return buf
	}

	t.Run("SVG", func(t *testing.T) {
		for _, testFilename := range []string{"kennedy.ged", "game_of_thrones.ged", "simpsons.ged"} {
			ctx := context.Background()
			pathToFile := filepath.Join("..", "..", "testdata", testFilename)
			buf := gedcomToMermaid(t, ctx, pathToFile)

			t.Run(testFilename, func(t *testing.T) { testSVG(t, ctx, buf) })
		}
	})

	t.Run("PNG", func(t *testing.T) {
		for _, testFilename := range []string{"kennedy.ged", "game_of_thrones.ged", "simpsons.ged"} {
			ctx := context.Background()
			pathToFile := filepath.Join("..", "..", "testdata", testFilename)
			buf := gedcomToMermaid(t, ctx, pathToFile)

			t.Run(testFilename, func(t *testing.T) {
				for _, scale := range []float64{1, 5, 10} {
					name := strconv.FormatFloat(scale, 'f', 1, 64)

					t.Run(name, func(t *testing.T) {
						testPNG(t, ctx, buf, scale)
					})
				}
			})
		}
	})
}

func testSVG(t *testing.T, ctx context.Context, r io.Reader) {
	t.Helper()

	m, err := NewMermaidRenderer(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cerr := m.Close(); cerr != nil {
			t.Log(cerr)
		}
	}()

	buf := new(bytes.Buffer)
	if err = m.DrawSVG(ctx, buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	const expPrefix = "<svg"
	if !strings.HasPrefix(got, expPrefix) {
		t.Fatalf("expected SVG output to begin with %q", expPrefix)
	}
	const expSuffix = "</svg>"
	if !strings.HasSuffix(got, expSuffix) {
		t.Fatalf("expected SVG output to end with %q", expSuffix)
	}
}

func testPNG(t *testing.T, ctx context.Context, r io.Reader, scale float64) {
	t.Helper()

	m, err := NewMermaidRenderer(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cerr := m.Close(); cerr != nil {
			t.Log(cerr)
		}
	}()

	buf := new(bytes.Buffer)
	if err = m.DrawPNG(ctx, buf, scale); err != nil {
		t.Fatal(err)
	}

	if _, err = png.Decode(buf); err != nil {
		t.Fatal(err)
	}
}
