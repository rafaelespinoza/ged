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
	out := alf.Command{
		Description: "read GEDCOM data from STDIN, transform to entity types, write to STDOUT",
		Setup: func(_ flag.FlagSet) *flag.FlagSet {
			fullName := mainName + " " + name
			flags := newFlagSet(fullName)

			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input

Description:
	Pipe in some data, interpret it and print the transformed results to STDOUT as JSON.

	This subcommand is for meant for inspecting the data transformations applied
	to the input data.

	The output shape:
		{
		  "people": []entity.Person{}.
		  "unions": []entity.Union{},
		}
`,
					initUsageLine(name),
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) error {
			people, unions, err := srv.ParseGedcom(context.Background(), os.Stdin)
			if err != nil {
				return err
			}

			return writeJSON(os.Stdout, map[string]any{"people": people, "unions": unions})
		},
	}

	return &out
}
