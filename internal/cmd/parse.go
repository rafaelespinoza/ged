package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rafaelespinoza/alf"
	"github.com/rafaelespinoza/ged/internal/gedcom"
	"github.com/rafaelespinoza/ged/internal/srv"
)

func makeParse(name string) alf.Directive {
	toEntities := alf.Command{
		Description: "transform data to application entity types",
		Setup: func(_ flag.FlagSet) *flag.FlagSet {
			subName := "to-entities"
			fullName := mainName + " " + subName
			flags := newFlagSet(fullName)

			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input

Description:
	Pipe in some data, interpret it and print the transformed results to STDOUT as JSON. The
	output shapes here are application "entities", which are simplified reductions of the
	original GEDCOM data. Most application functionality interfaces with these data types.

	The output shape:
		{
		  "people": []entity.Person{},
		  "unions": []entity.Union{}
		}
`,
					initUsageLine(subName),
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

	toRecords := alf.Command{
		Description: "transform data to GEDCOM records",
		Setup: func(_ flag.FlagSet) *flag.FlagSet {
			subName := "to-records"
			fullName := mainName + " " + subName
			flags := newFlagSet(fullName)

			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input

Description:
	Pipe in some data, interpret it and print the transformed results to STDOUT as JSON.
	The output shapes are like GEDCOM record types.

	The output shape:
		{
		  "Individuals": []gedcom.IndividualRecord{},
		  "Families":    []gedcom.FamilyRecord{},
		  "Sources":     []gedcom.SourceRecord{}
		}
`,
					initUsageLine(subName),
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) error {
			records, err := gedcom.ReadRecords(ctx, os.Stdin)
			if err != nil {
				return err
			}

			return writeJSON(os.Stdout, records)
		},
	}

	out := alf.Delegator{
		Description: "interpret GEDCOM data, transform it, write to STDOUT",
		Subs: map[string]alf.Directive{
			"to-entities": &toEntities,
			"to-records":  &toRecords,
		},
		Flags: newFlagSet(mainName),
	}
	out.Flags.Usage = func() {
		fmt.Fprintf(out.Flags.Output(), `%s < path/to/input

Description:
	Pipe in some data, interpret it and print the transformed results to STDOUT as JSON.

	This subcommand is for meant for inspecting the data transformations applied
	to the input data.

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v
`,
			initUsageLine(name), strings.Join(out.DescribeSubcommands(), "\n\t"),
		)
	}

	return &out
}
