package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rafaelespinoza/alf"
	"github.com/rafaelespinoza/ged/internal/entity"
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

	toLines := alf.Command{
		Description: "transform data to a single line format for fzf processing",
		Setup: func(_ flag.FlagSet) *flag.FlagSet {
			subName := "to-lines"
			fullName := mainName + " " + subName
			flags := newFlagSet(fullName)

			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input

Description:
	This subcommand formats the GEDCOM data for some application-specific workflows.
	Its output is meant as input for fzf, a very awesome fuzzy-finder.

	The output shape is one individual record per line. Within each line, each
	field is delimited by one ASCII TAB (%q, 0x%x). If a field value is empty,
	then that key-value pair is omitted.

	Example output lines:

	%s
	%s
`,
					initUsageLine(subName),
					fzfLineFieldSeparator, fzfLineFieldSeparator,
					strings.Join([]string{"@I123@", "Name:Full Name of Person", "Birth:2006-01-02", "Death:2038-01-17", "Parent:Name Of Parent1", "Parent:Name of Parent2"}, fzfLineFieldSeparator),
					strings.Join([]string{"@I234@", "Name:Full Name of Person", "Birth:2006-01-02", "Parent:Name Of Parent"}, fzfLineFieldSeparator),
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) error {
			people, _, err := srv.ParseGedcom(context.Background(), os.Stdin)
			if err != nil {
				return err
			}
			for _, line := range makeFZFPeople(fzfLineFieldSeparator, people) {
				if _, err = fmt.Println(line); err != nil {
					return err
				}
			}
			return nil
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
			"to-lines":    &toLines,
			"to-records":  &toRecords,
		},
		Flags: newFlagSet(mainName),
	}
	out.Flags.Usage = func() {
		fmt.Fprintf(out.Flags.Output(), `%s < path/to/input

Description:
	Pipe in some GEDCOM data, interpret it and print the transformed results to STDOUT.


	This subcommand is for meant for inspecting the data transformations applied
	to the input data, or preparing it for further processing.

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v
`,
			initUsageLine(name), strings.Join(out.DescribeSubcommands(), "\n\t"),
		)
	}

	return &out
}

const fzfLineFieldSeparator = "\t"

func makeFZFPeople(fieldDelimiter string, people []*entity.Person) (out []string) {
	out = make([]string, len(people))
	var line []string

	for i, person := range people {
		// Set cap to maximum number of non-empty fields you might have (assuming 2 parents).
		line = make([]string, 0, 6)

		line = append(line, person.ID)
		line = append(line, "Name:"+person.Name.Full())
		if d := formatDate(person.Birthdate); d != nil {
			line = append(line, "Birth:"+*d)
		}
		if d := formatDate(person.Deathdate); d != nil {
			line = append(line, "Death:"+*d)
		}
		for _, parent := range person.Parents {
			line = append(line, "Parent:"+parent.Name.Full())
		}

		out[i] = strings.Join(line, fieldDelimiter)
	}

	return
}

func formatDate(in *entity.Date) *string {
	if in == nil {
		return nil
	}

	out := in.String()
	return &out
}
