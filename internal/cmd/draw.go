package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rafaelespinoza/alf"
	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/srv"
)

func makeDraw(name string) alf.Directive {
	var inputFormat, flowchartDirection string
	var displayID bool
	supportedInputFormats := []string{"gedcom", "json"}
	out := alf.Command{
		Description: "generate a family tree",
		Setup: func(p flag.FlagSet) *flag.FlagSet {
			fullName := mainName + " " + name
			flags := newFlagSet(fullName)
			flags.StringVar(&flowchartDirection, "direction", srv.ValidFlowchartDirections[0], fmt.Sprintf("orientation of flowchart, one of %q", srv.ValidFlowchartDirections))
			flags.StringVar(&inputFormat, "f", supportedInputFormats[0], fmt.Sprintf("input format, one of %q", supportedInputFormats))
			flags.BoolVar(&displayID, "display-id", false, "if true, then display ID of each person")

			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input

Description:
	Pipe in some data, draw a family tree as a mermaid flowchart to STDOUT.

	The format of the input data could be "gedcom".

	If the format is "json", then the shape of that data should be:
		{
		  "people": []entity.Person{}
		  "unions": []entity.Union{}
		}
	See the parse subcommand to format that JSON.

	If display-id is true, then each person node will have the ID of the person displayed.
	This ID can be helpful for additional introspection.

Examples:
	# Using gedcom-formatted input
	$ %s < path/to/data.ged

	# Using json-formatted input.
	$ %s -f json < path/to/data.json
	$ jq '.' path/to/data.json | %s -f json
`,
					initUsageLine(name), fullName, fullName, fullName,
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) (err error) {
			var (
				people []*entity.Person
				unions []*entity.Union
			)

			switch inputFormat {
			case "json":
				var jsonData struct {
					People []*entity.Person
					Unions []*entity.Union
				}
				if err = readJSON(os.Stdin, &jsonData); err != nil {
					return
				}
				people, unions = jsonData.People, jsonData.Unions
			default:
				if people, unions, err = srv.ParseGedcom(ctx, os.Stdin); err != nil {
					return
				}
			}

			return srv.MakeMermaidFlowchart(ctx, srv.MermaidFlowchartParams{
				Direction: flowchartDirection,
				DisplayID: displayID,
				Out:       os.Stdout,
				People:    people,
				Unions:    unions,
			})
		},
	}

	return &out
}
