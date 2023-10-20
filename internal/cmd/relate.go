package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rafaelespinoza/alf"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/log"
	"github.com/rafaelespinoza/ged/internal/srv"
)

func makeRelate(name string) alf.Directive {
	var inputFormat string
	supportedInputFormats := []string{"gedcom", "json"}
	out := alf.Command{
		Setup: func(p flag.FlagSet) *flag.FlagSet {
			name := bin + " " + name
			flags := newFlagSet(name)
			flags.StringVar(&inputFormat, "f", supportedInputFormats[0], fmt.Sprintf("input format, one of %q", supportedInputFormats))
			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `Usage: %s [flags] < path/to/input

Description:
	Pipe in some data, calculate relationship between 2 people.
	The input data should represent an array of people ([]entity.Person).

	The format of the input data could be gedcom.

	Or it could be json; see the parse subcommand format that json.
	If it's json, then remember to pipe it array of entity.People.

Examples:
	# Using gedcom-formatted input
	$ %s < path/to/data.ged

	# Using json-formatted input. Remember to pipe it an array of People structs.
	$ %s -f json < <(jq '.people' < path/to/data.json)
	$ jq '.people' < path/to/data.json | %s -f json
`,
					name, name, name, name,
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) (err error) {
			var (
				people     []*entity.Person
				p1ID, p2ID string
				r1, r2     entity.Lineage
			)
			switch inputFormat {
			case "json":
				if err = readJSON(os.Stdin, &people); err != nil {
					return
				}
			default:
				if people, _, err = srv.ParseGedcom(ctx, os.Stdin); err != nil {
					return
				}
			}

			if p1ID, p2ID, err = choosePeopleToRelate(ctx, people); err != nil {
				return
			}

			if r1, r2, err = srv.NewRelator(people).Relate(ctx, p1ID, p2ID); err != nil {
				return
			}

			for _, rel := range []entity.Lineage{r1, r2} {
				commonAncestors := make([]map[string]any, len(rel.CommonAncestors))

				for j, person := range rel.CommonAncestors {
					commonAncestors[j] = map[string]any{
						"id":         person.ID,
						"name":       person.Name,
						"birth_date": formatDate(person.Birthdate),
						"death_date": formatDate(person.Deathdate),
					}
				}

				err = writeJSON(os.Stdout, map[string]any{
					"description":         rel.Description,
					"type":                rel.Type.String(),
					"generations_removed": rel.GenerationsRemoved,
					"common_ancestors":    commonAncestors,
				})
				if err != nil {
					return
				}
			}

			return
		},
	}

	return &out
}

func choosePeopleToRelate(ctx context.Context, people []*entity.Person) (p1ID, p2ID string, err error) {
	inputLines := make([]string, len(people))
	var inputLine []string
	for i, person := range people {
		// Set cap to maximum number of non-empty fields you might have (assuming 2 parents).
		inputLine = make([]string, 0, 6)

		if person.ID != "" {
			inputLine = append(inputLine, "ID:"+person.ID)
		}
		inputLine = append(inputLine, "Name:"+person.Name.Full())
		if d := formatDate(person.Birthdate); d != nil {
			inputLine = append(inputLine, "Birth:"+*d)
		}
		if d := formatDate(person.Deathdate); d != nil {
			inputLine = append(inputLine, "Death:"+*d)
		}
		for _, parent := range person.Parents {
			inputLine = append(inputLine, "Parent:"+parent.Name.Full())
		}

		inputLines[i] = strings.Join(inputLine, "\t")
	}

	p1ID, err = fuzzyChooseID(ctx, inputLines)
	if err != nil {
		err = fmt.Errorf("error on choice 1: %w", err)
		return
	}

	p2ID, err = fuzzyChooseID(ctx, inputLines)
	if err != nil {
		err = fmt.Errorf("error on choice 2: %w", err)
		return
	}

	return
}

func fuzzyChooseID(ctx context.Context, inputLines []string) (out string, err error) {
	const fzfOptions = "--header 'pick person to compare (try fuzzy search)' --layout reverse --height 66% --info=inline --border --margin=1 --padding=1 --cycle"

	filtered := withFilter(ctx, "fzf "+fzfOptions, func(in io.WriteCloser) {
		for _, line := range inputLines {
			fmt.Fprintln(in, line)
		}
	})

	// parsing the user choice depends highly on the formatting done to the
	// input line.
	var userChoice string
	if len(filtered) < 1 {
		err = errors.New("no lines chosen")
		return
	} else if len(filtered) > 1 {
		userChoice = filtered[0]
	}
	fields := strings.Fields(userChoice)
	if len(fields) < 1 {
		err = errors.New("it seems like an early exit")
		return
	}

	fieldParts := strings.Split(fields[0], ":")
	if len(fieldParts) < 2 {
		err = fmt.Errorf("expected first field to have %q, field=%q", ":", fields[0])
		return
	}

	out = fieldParts[1]
	return
}

// withFilter was lifted from https://github.com/junegunn/fzf/wiki/Language-bindings#go.
func withFilter(ctx context.Context, command string, handleStdin func(in io.WriteCloser)) []string {
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}

	cmd := exec.CommandContext(ctx, shell, "-c", command)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Error(ctx, nil, err, "stdin error")
	}

	go func() {
		handleStdin(stdin)
		if cerr := stdin.Close(); cerr != nil {
			log.Error(ctx, nil, err, "could not close input pipe")
		}
	}()

	result, err := cmd.Output()
	if err != nil {
		log.Error(ctx, nil, err, "stdout error")
	}
	return strings.Split(string(result), "\n")
}

func formatDate(in *time.Time) *string {
	if in == nil {
		return nil
	}

	val := in.Format(time.DateOnly)
	return &val
}
