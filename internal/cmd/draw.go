package cmd

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/rafaelespinoza/alf"
	"github.com/rafaelespinoza/ged/internal/log"
	"github.com/rafaelespinoza/ged/internal/srv"
)

func makeDraw(name string) alf.Directive {
	var flowchartDirection, inputFormat, outputFormat string
	var displayID bool
	var renderPNGScale float64

	const mermaid = "mermaid"

	supportedInputFormats := []string{"gedcom", mermaid}
	supportedOutputFormats := []string{"svg", "png", mermaid}

	fullName := mainName + " " + name
	flags := newFlagSet(fullName)

	out := alf.Command{
		Description: "generate a family tree",
		Setup: func(p flag.FlagSet) *flag.FlagSet {
			flags.StringVar(&flowchartDirection, "direction", srv.ValidFlowchartDirections[0], fmt.Sprintf("orientation of flowchart, one of %q", srv.ValidFlowchartDirections))
			flags.BoolVar(&displayID, "display-id", false, "show each person's ID in flowchart")
			flags.StringVar(&inputFormat, "input-format", supportedInputFormats[0], fmt.Sprintf("format of the input data, one of %q", supportedInputFormats))

			flags.StringVar(&outputFormat, "output-format", supportedOutputFormats[0], fmt.Sprintf("format of output data, one of %q", supportedOutputFormats))
			flags.Float64Var(&renderPNGScale, "render-png-scale", 10.0, "scaling factor for rendering PNG")

			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input

Description:
	Pipe in some data from STDIN, draw a family tree to STDOUT.

	This subcommand can:
		- take in GEDCOM data and output a Mermaid-formatted flowchart
		- take in a Mermaid-formatted flowchart and output a Mermaid-rendered SVG or PNG
		- do both in one fell swoop: take in GEDCOM data and output an SVG or PNG

Flowchart-related options:
	Set the flowchart's direction with the flag, direction.

	If display-id is true, then each person node will have the ID of the person displayed.
	This ID can be helpful for additional introspection.

Input-related options:
	If you'd prefer to make some manual edits to the Mermaid flowchart, and you
	want to render it again, just specify -input-format=%s.
	Otherwise, the input is assumed to be GEDCOM-formatted data.

Output-related options:
	Most of the time, you probably just want to go from GEDCOM data directly to
	an SVG or PNG. You also have the option of outputting the Mermaid flowchart,
	which can be useful if you'd prefer to make your own edits.
	Set output-format=%s, to do that.

	For outputting PNG, specify a scaling factor with the flag render-png-scale.
	If the flowchart direction is vertical (like TD, BT, TB), then you may want
	to consider a higher scaling factor. But if the flowchart direction is
	horizontal (like LR, RL), then a lower scaling factor is usually sufficient.

Examples:
	# Interpret GEDCOM data and render SVG
	$ %s < path/to/data.ged

	# Interpret GEDCOM data and generate a Mermaid flowchart.
	# Then render the Mermaid flowchart, this time as PNG.
	$ %s -output-format=%s < path/to/data.ged > path/to/data.mermaid
	$ %s -input-format=%s -output-format=png < path/to/data.mermaid > path/to/data.png
`,
					initUsageLine(name), mermaid, mermaid,
					fullName,
					fullName, mermaid,
					fullName, mermaid,
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) (err error) {
			if !slices.Contains(supportedInputFormats, inputFormat) {
				return fmt.Errorf("invalid input-format %q, valid ones are: %q", inputFormat, supportedInputFormats)
			}
			if !slices.Contains(supportedOutputFormats, outputFormat) {
				return fmt.Errorf("invalid output-format %q, valid ones are: %q", outputFormat, supportedOutputFormats)
			}

			if inputFormat == mermaid && outputFormat == mermaid {
				err = errors.New("invalid combination of -input-format and -output-format")
				return
			} else if outputFormat == mermaid {
				// take in GEDCOM data and generate a Mermaid flowchart
				err = makeMermaidFlowchart(ctx, os.Stdin, os.Stdout, flowchartDirection, displayID)
				return
			} else if inputFormat == mermaid {
				// take in a Mermaid flowchart, and render it (via Mermaid) as SVG or PNG
				err = renderMermaidFlowchart(ctx, os.Stdin, os.Stdout, outputFormat, renderPNGScale)
				return
			}

			// take in GEDCOM data, then generate a Mermaid flowchart, then
			// render that flowchart as SVG or PNG.

			chartIO := new(bytes.Buffer)
			if err = makeMermaidFlowchart(ctx, os.Stdin, chartIO, flowchartDirection, displayID); err != nil {
				return
			}

			err = renderMermaidFlowchart(ctx, chartIO, os.Stdout, outputFormat, renderPNGScale)
			return
		},
	}

	return &out
}

func makeMermaidFlowchart(ctx context.Context, r io.Reader, w io.Writer, flowchartDirection string, displayID bool) (err error) {
	people, unions, err := srv.ParseGedcom(ctx, r)
	if err != nil {
		return
	}

	err = srv.MakeMermaidFlowchart(ctx, srv.MermaidFlowchartParams{
		Direction: flowchartDirection,
		DisplayID: displayID,
		Out:       w,
		People:    people,
		Unions:    unions,
	})

	return
}

func renderMermaidFlowchart(ctx context.Context, r io.Reader, w io.Writer, renderFormat string, pngScale float64) (err error) {
	m, err := srv.NewMermaidRenderer(ctx, r)
	if err != nil {
		return
	}

	defer func() {
		if cerr := m.Close(); cerr != nil {
			log.Error(ctx, nil, cerr, "failed to close mermaid")
		}
	}()

	switch renderFormat {
	case "png":
		err = m.DrawPNG(ctx, w, pngScale)
	default:
		err = m.DrawSVG(ctx, w)
	}

	return
}
