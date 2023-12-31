package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/rafaelespinoza/alf"
	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/gedcom"
)

type viewGroupSheetInputs struct {
	peopleByID   map[string]*gedcom.IndividualRecord
	familiesByID map[string]*gedcom.FamilyRecord
	targetID     string
}

func makeExploreDataShow(parentName, name string) alf.Directive {
	var showParams viewGroupSheetInputs
	var outputFormat string
	supportedOutputFormats := []string{"", "json"}
	out := alf.Command{
		Description: "display transformed GEDCOM data in a group sheet view",
		Setup: func(_ flag.FlagSet) *flag.FlagSet {
			fullName := strings.Join([]string{mainName, parentName, name}, " ")
			flags := newFlagSet(fullName)

			flags.StringVar(&showParams.targetID, "target-id", "", "GEDCOM Xref to display")
			flags.StringVar(&outputFormat, "output-format", supportedOutputFormats[0], fmt.Sprintf("output format, one of %q", supportedOutputFormats))
			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input.ged

Description:
`,
					initUsageLine(name),
				)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) error {
			if showParams.targetID == "" {
				return errors.New("target-id is required")
			}

			records, err := gedcom.ReadRecords(ctx, os.Stdin)
			if err != nil {
				return err
			}

			showParams.peopleByID = make(map[string]*gedcom.IndividualRecord, len(records.Individuals))
			for _, individual := range records.Individuals {
				showParams.peopleByID[individual.Xref] = individual
			}
			showParams.familiesByID = make(map[string]*gedcom.FamilyRecord, len(records.Families))
			for _, fam := range records.Families {
				showParams.familiesByID[fam.Xref] = fam
			}

			data, err := buildGroupSheetView(showParams)
			if err != nil {
				return err
			}

			switch outputFormat {
			case supportedOutputFormats[1]:
				err = writeJSON(os.Stdout, data)
			default:
				err = renderGroupSheetView(os.Stdout, data)
			}
			return err
		},
	}

	return &out
}

func buildGroupSheetView(in viewGroupSheetInputs) (*groupSheetView, error) {
	target, ok := in.peopleByID[in.targetID]
	if !ok {
		return nil, fmt.Errorf("person with ID %q not found", in.targetID)
	}

	var (
		out groupSheetView
		err error
	)

	out.Person = buildGroupSheetPerson(target)
	out.Notes = make([]string, len(target.Notes))
	out.FamiliesAsChild = make([]groupSheetFamily, len(target.FamiliesAsChild))
	out.FamiliesAsPartner = make([]groupSheetFamily, len(target.FamiliesAsPartner))

	for i, note := range target.Notes {
		out.Notes[i] = note.Payload
	}
	for i, famID := range target.FamiliesAsChild {
		out.FamiliesAsChild[i], err = buildGroupSheetFamily(famID, in.peopleByID, in.familiesByID)
		if err != nil {
			return nil, fmt.Errorf("could not make families as child: %w", err)
		}
	}
	for i, famID := range target.FamiliesAsPartner {
		out.FamiliesAsPartner[i], err = buildGroupSheetFamily(famID, in.peopleByID, in.familiesByID)
		if err != nil {
			return nil, fmt.Errorf("could not make families as partner: %w", err)
		}
	}

	return &out, nil
}

func buildGroupSheetPerson(in *gedcom.IndividualRecord) *groupSheetSimplePerson {
	var name string
	if len(in.Names) > 0 {
		name = in.Names[0].String()
	}
	var birth, death groupSheetDate
	if len(in.Birth) > 0 {
		birth = buildGroupSheetDate(in.Birth[0])
	}
	if len(in.Death) > 0 {
		death = buildGroupSheetDate(in.Death[0])
	}

	return &groupSheetSimplePerson{
		ID:    in.Xref,
		Name:  name,
		Birth: birth,
		Death: death,
	}
}

func buildGroupSheetDate(in *gedcom.Event) (out groupSheetDate) {
	if in == nil {
		return
	}
	date, _ := entity.NewDate(in.Date, in.DateRange)
	out = groupSheetDate{
		Date:  date.String(),
		Place: in.Place,
	}
	return
}

func buildGroupSheetFamily(famID string, peopleByID map[string]*gedcom.IndividualRecord, familiesByID map[string]*gedcom.FamilyRecord) (out groupSheetFamily, err error) {
	fam, ok := familiesByID[famID]
	if !ok {
		err = fmt.Errorf("family with ID %q not found", famID)
		return
	}

	parents := make([]*groupSheetSimplePerson, len(fam.ParentXrefs))
	for j, parentID := range fam.ParentXrefs {
		individual, ok := peopleByID[parentID]
		if !ok {
			err = fmt.Errorf("parent with ID %q not found", parentID)
			return
		}
		parents[j] = buildGroupSheetPerson(individual)
		parents[j].Role = "parent"
	}

	children := make([]*groupSheetSimplePerson, 0, len(fam.ChildXrefs))
	for _, childID := range fam.ChildXrefs {
		individual, ok := peopleByID[childID]
		if !ok {
			err = fmt.Errorf("child with ID %q not found", childID)
			return
		}
		children = append(children, buildGroupSheetPerson(individual))
		children[len(children)-1].Role = "child"
	}

	out = groupSheetFamily{
		ID:         famID,
		MarriedAt:  buildGroupSheetDate(fam.MarriedAt),
		DivorcedAt: buildGroupSheetDate(fam.DivorcedAt),
		Parents:    parents,
		Children:   slices.Clip(children),
	}

	return
}
