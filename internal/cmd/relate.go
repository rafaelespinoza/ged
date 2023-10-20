package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	var inputFormat, p1ID, p2ID string
	supportedInputFormats := []string{"gedcom", "json"}
	out := alf.Command{
		Description: "calculate, describe relationship between people",
		Setup: func(p flag.FlagSet) *flag.FlagSet {
			fullname := bin + " " + name
			flags := newFlagSet(fullname)
			flags.StringVar(&inputFormat, "f", supportedInputFormats[0], fmt.Sprintf("input format, one of %q", supportedInputFormats))
			flags.StringVar(&p1ID, "p1", "", "id of person 1")
			flags.StringVar(&p2ID, "p2", "", "id of person 2")

			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `Usage: %s [ged-flags] %s [%s-flags] < path/to/input

Description:
	Pipe in some input data, calculate relationship between 2 people.

Format of input data:
	The input data should represent an array of people ([]entity.Person).

	The format of the input data could be "gedcom".

	If the format is "json", then the shape of that data should be:
		{
		  "people": []entity.Person{}
		}
	See the parse subcommand to format that JSON.

Choosing people to relate:
	You must also select 2 people for which to calculate the relationship by
	specifying their IDs. There are multiple ways to do this.

	Directly, via flags -p1, -p2:
		If you know both IDs ahead of time, you can input them directly via the
		flags -p1 and p2. If you go this route, be sure to specify both values.

	Via "fzf":
		If you'd prefer to discover their IDs, you can do a fuzzy search.
		This requires fzf, a really awesome fuzzy finder for the command line.
		Check it out at:
			https://github.com/junegunn/fzf
		Once fzf is available, be sure to omit the flag values for -p1 and -p2.
		The binary for fzf does not necessarily need to be in your PATH. You can
		specify the path to fzf via the env var FZF_BIN.

Examples:
	# Using gedcom-formatted data, use a fuzzy search to choose people to relate.
	$ %s < path/to/data.ged

	# Using json-formatted data.
	$ %s -f json < path/to/data.json
	$ jq '.' path/to/data.json | %s -f json

	# Using gedcom-formatted data. Directly input person IDs.
	$ %s -p1 @I111@ -p2 @I222@ < path/to/data.ged

	# Using json-formatted data. Directly input person IDs.
	$ %s -f json -p1 @I111@ -p2 @I222@ < path/to/data.json
`,
					bin, name, name, fullname, fullname, fullname, fullname, fullname,
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) (err error) {
			var (
				people []*entity.Person
				r1, r2 entity.Lineage
			)
			switch inputFormat {
			case "json":
				var data struct{ People []*entity.Person }
				if err = readJSON(os.Stdin, &data); err != nil {
					return
				}
				people = data.People
			default:
				if people, _, err = srv.ParseGedcom(ctx, os.Stdin); err != nil {
					return
				}
			}

			if p1ID != "" && p2ID == "" || p1ID == "" && p2ID != "" {
				err = fmt.Errorf("presence of flag values for -p1 (%q) and -p2 (%q) are mutually-exclusive; pass in values for both flags or omit both", p1ID, p2ID)
				return
			} else if p1ID == "" && p2ID == "" {
				if p1ID, p2ID, err = choosePeopleToRelate(ctx, people); err != nil {
					return
				}
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
	chooser, err := newFZFChooser(ctx, people)
	if errors.Is(err, errNoFZF) {
		log.Error(ctx, nil, err, "install fzf for a better user experience; see https://github.com/junegunn/fzf")
	}
	if err != nil {
		return
	}

	p1ID, err = chooser.choosePersonID(ctx)
	if err != nil {
		err = fmt.Errorf("error on choice 1: %w", err)
		return
	}

	p2ID, err = chooser.choosePersonID(ctx)
	if err != nil {
		err = fmt.Errorf("error on choice 2: %w", err)
		return
	}
	return
}

type (
	personIDChooser interface {
		choosePersonID(ctx context.Context) (id string, err error)
	}

	fzfChooser struct {
		bin        string
		inputLines []string
	}
)

var errNoFZF = errors.New("fzf binary not available")

func newFZFChooser(ctx context.Context, people []*entity.Person) (personIDChooser, error) {
	fzf := os.Getenv("FZF_BIN")
	if len(fzf) == 0 {
		fzf = "fzf"
	}
	pathToBin, err := exec.LookPath(fzf)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errNoFZF, err)
	}

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

	return &fzfChooser{bin: pathToBin, inputLines: inputLines}, nil
}

func (c *fzfChooser) choosePersonID(ctx context.Context) (out string, err error) {
	// this part was adapted from https://github.com/junegunn/fzf/wiki/Language-bindings#go.
	const fzfOptions = "--header 'pick person to compare (try fuzzy search)' --layout reverse --height 66% --info=inline --border --margin=1 --padding=1 --cycle"
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}

	cmd := exec.CommandContext(ctx, shell, "-c", c.bin+" "+fzfOptions)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Error(ctx, nil, err, "stdin error")
	}
	go func() {
		for _, line := range c.inputLines {
			fmt.Fprintln(stdin, line)
		}
		if cerr := stdin.Close(); cerr != nil {
			log.Error(ctx, nil, err, "could not close input pipe")
		}
	}()

	var userChoice string
	if result, rerr := cmd.Output(); rerr != nil {
		err = rerr
		return
	} else {
		before, after, found := strings.Cut(string(result), "\n")
		log.Info(ctx, map[string]any{"before": before, "after": after, "found": found}, "")
		userChoice = before
	}

	// parsing the user choice depends highly on the formatting done to the
	// input line.
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

func formatDate(in *time.Time) *string {
	if in == nil {
		return nil
	}

	val := in.Format(time.DateOnly)
	return &val
}
