package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rafaelespinoza/alf"
	"github.com/rafaelespinoza/ged/internal/srv"
)

func makeParse(name string) alf.Directive {
	var inputFormat string
	supportedInputFormats := []string{"gedcom"}

	out := alf.Command{
		Description: "read GEDCOM data from STDIN, transform to entity types, write to STDOUT",
		Setup: func(_ flag.FlagSet) *flag.FlagSet {
			flags := newFlagSet(name)
			flags.StringVar(&inputFormat, "f", "gedcom", fmt.Sprintf("input format, one of %q", supportedInputFormats))
			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `Usage: %s [flags] < path/to/input

Description:
	Pipe in some data, interpret it and print the transformed results to STDOUT as JSON.

	This subcommand is for meant for inspecting the data transformations applied
	to the input data.

	Supported input formats:
		%q

	The output shape:
		{
			"people": []entity.Person{}.
			"unions": []entity.Union{},
		}
`,
					name, supportedInputFormats,
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) error {
			switch inputFormat {
			case supportedInputFormats[0]:
				break
			default:
				return fmt.Errorf("invalid input format (flag -f) %q", inputFormat)
			}

			people, unions, err := srv.ParseGedcom(context.Background(), os.Stdin)
			if err != nil {
				return err
			}

			return writeJSON(os.Stdout, map[string]any{"people": people, "unions": unions})
		},
	}

	return &out
}
