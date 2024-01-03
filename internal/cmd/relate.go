package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/rafaelespinoza/alf"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/log"
	"github.com/rafaelespinoza/ged/internal/srv"
)

func makeRelate(name string) alf.Directive {
	var inputFormat, outputFormat, p1ID, p2ID string
	supportedInputFormats := []string{"gedcom", "json"}
	supportedOutputFormats := []string{"", "json"}
	out := alf.Command{
		Description: "calculate, describe relationship between people",
		Setup: func(p flag.FlagSet) *flag.FlagSet {
			fullName := mainName + " " + name
			flags := newFlagSet(fullName)
			flags.StringVar(&inputFormat, "f", supportedInputFormats[0], fmt.Sprintf("input format, one of %q", supportedInputFormats))
			flags.StringVar(&outputFormat, "output-format", supportedOutputFormats[0], fmt.Sprintf("output format, one of %q", supportedOutputFormats))
			flags.StringVar(&p1ID, "p1", "", "id of person 1")
			flags.StringVar(&p2ID, "p2", "", "id of person 2")

			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input

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
					initUsageLine(name), fullName, fullName, fullName, fullName, fullName,
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) (err error) {
			var (
				people []*entity.Person
				result entity.MutualRelationship
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

			if result, err = srv.NewRelator(people).Relate(ctx, p1ID, p2ID); err != nil {
				return
			}

			mr := makeMutualRelationship(result)

			switch outputFormat {
			case supportedOutputFormats[1]:
				err = writeJSON(os.Stdout, mr)
			default:
				err = formatMutualRelationship(os.Stdout, mr)
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

type personIDChooser interface {
	choosePersonID(ctx context.Context) (id string, err error)
}

func formatDate(in *entity.Date) *string {
	if in == nil {
		return nil
	}

	out := in.String()
	return &out
}

type (
	mutualRelationship struct {
		Person1       *groupSheetSimplePerson   `json:"person_1"`
		Person2       *groupSheetSimplePerson   `json:"person_2"`
		Union         []*groupSheetSimplePerson `json:"union"`
		CommonPerson  *groupSheetSimplePerson   `json:"common_person"`
		Relationship1 *relationship             `json:"relationship_1"`
		Relationship2 *relationship             `json:"relationship_2"`
	}
	relationship struct {
		Description        string                    `json:"description"`
		Type               string                    `json:"type"`
		GenerationsRemoved int                       `json:"generations_removed"`
		Path               []*groupSheetSimplePerson `json:"path"`
		SourceID           string                    `json:"source_id"`
		TargetID           string                    `json:"target_id"`
	}
)

func makeMutualRelationship(in entity.MutualRelationship) (out mutualRelationship) {
	if in.CommonPerson != nil {
		out.CommonPerson = simplifyPerson(*in.CommonPerson)
	}

	if in.Union != nil {
		var p1, p2 *groupSheetSimplePerson
		if in.Union.Person1 != nil {
			p1 = simplifyPerson(*in.Union.Person1)
		}
		if in.Union.Person2 != nil {
			p2 = simplifyPerson(*in.Union.Person2)
		}

		out.Union = []*groupSheetSimplePerson{p1, p2}
	}

	type Tuple struct {
		Src        entity.Relationship
		RelDest    **relationship
		PersonDest **groupSheetSimplePerson
	}

	for _, tup := range []Tuple{{in.R1, &out.Relationship1, &out.Person1}, {in.R2, &out.Relationship2, &out.Person2}} {
		path := make([]*groupSheetSimplePerson, len(tup.Src.Path))
		for j, person := range tup.Src.Path {
			path[j] = simplifyPerson(person)

			if j == 0 {
				*tup.PersonDest = simplifyPerson(person)
			}
		}

		*tup.RelDest = &relationship{
			Description:        tup.Src.Description,
			Type:               tup.Src.Type.String(),
			GenerationsRemoved: tup.Src.GenerationsRemoved,
			Path:               path,
			SourceID:           tup.Src.SourceID,
			TargetID:           tup.Src.TargetID,
		}
	}

	return
}

func simplifyPerson(p entity.Person) *groupSheetSimplePerson {
	var birth, death groupSheetDate
	if p.Birthdate != nil {
		birth = groupSheetDate{Date: p.Birthdate.String()}
	}
	if p.Deathdate != nil {
		death = groupSheetDate{Date: p.Deathdate.String()}
	}
	return &groupSheetSimplePerson{
		ID:    p.ID,
		Role:  "",
		Name:  p.Name.Full(),
		Birth: birth,
		Death: death,
	}
}
