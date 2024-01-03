package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/rafaelespinoza/alf"
	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/gedcom"
)

func makeExploreData(name string) alf.Directive {
	var showParams viewGroupSheetInputs
	var outputFormat string
	supportedOutputFormats := []string{"", "json"}
	show := alf.Command{
		Description: "display transformed GEDCOM data in a group sheet view",
		Setup: func(_ flag.FlagSet) *flag.FlagSet {
			subName := "show"
			fullName := mainName + " " + name + " " + subName
			flags := newFlagSet(fullName)

			flags.StringVar(&showParams.targetID, "target-id", "", "GEDCOM Xref to display")
			flags.StringVar(&outputFormat, "output-format", supportedOutputFormats[0], fmt.Sprintf("output format, one of %q", supportedOutputFormats))
			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `%s < path/to/input.ged

Description:
`,
					initUsageLine(subName),
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

			data, err := makeGroupSheetView(showParams)
			if err != nil {
				return err
			}

			switch outputFormat {
			case supportedOutputFormats[1]:
				err = writeJSON(os.Stdout, data)
			default:
				err = formatGroupSheetView(os.Stdout, data)
			}
			return err
		},
	}
	out := alf.Delegator{
		Description: "view GEDCOM data",
		Subs: map[string]alf.Directive{
			"show": &show,
		},
		Flags: newFlagSet(mainName),
	}
	out.Flags.Usage = func() {
		fmt.Fprintf(out.Flags.Output(), `%s < path/to/input

Description:

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v
`,
			initUsageLine(name), strings.Join(out.DescribeSubcommands(), "\n\t"),
		)
	}

	return &out
}

type (
	viewGroupSheetInputs struct {
		peopleByID   map[string]*gedcom.IndividualRecord
		familiesByID map[string]*gedcom.FamilyRecord
		targetID     string
	}
	groupSheetView struct {
		Person            *groupSheetSimplePerson `json:"person"`
		Notes             []string                `json:"notes"`
		FamiliesAsChild   []familiesAsChild       `json:"families_as_child"`
		FamiliesAsPartner []familiesAsPartner     `json:"families_as_partner"`
	}
	groupSheetSimplePerson struct {
		ID    string         `json:"id"`
		Role  string         `json:"role"`
		Name  string         `json:"name"`
		Birth groupSheetDate `json:"birth"`
		Death groupSheetDate `json:"death"`
	}
	groupSheetDate struct {
		Date  string `json:"date"`
		Place string `json:"place"`
	}
	familiesAsChild struct {
		ID       string                    `json:"id"`
		Parents  []*groupSheetSimplePerson `json:"parents"`
		Siblings []*groupSheetSimplePerson `json:"siblings"`
	}
	familiesAsPartner struct {
		ID       string                    `json:"id"`
		Partner  *groupSheetSimplePerson   `json:"partner"`
		Children []*groupSheetSimplePerson `json:"children"`
	}
)

func makeGroupSheetView(in viewGroupSheetInputs) (*groupSheetView, error) {
	target, ok := in.peopleByID[in.targetID]
	if !ok {
		return nil, fmt.Errorf("person with ID %q not found", in.targetID)
	}

	var (
		out groupSheetView
		err error
	)

	out.Person = makeGroupSheetPerson(target)
	out.Notes = make([]string, len(target.Notes))
	for i, note := range target.Notes {
		out.Notes[i] = note.Payload
	}

	out.FamiliesAsChild, err = makeFamiliesAsChild(in, target)
	if err != nil {
		return nil, fmt.Errorf("could not make families as child: %w", err)
	}

	out.FamiliesAsPartner, err = makeFamiliesAsPartner(in, target)
	if err != nil {
		return nil, fmt.Errorf("could not make families as partner: %w", err)
	}

	return &out, nil
}

func makeGroupSheetPerson(in *gedcom.IndividualRecord) *groupSheetSimplePerson {
	var name string
	if len(in.Names) > 0 {
		name = in.Names[0].String()
	}

	return &groupSheetSimplePerson{
		ID:    in.Xref,
		Name:  name,
		Birth: makeGroupSheetDate(in.Birth),
		Death: makeGroupSheetDate(in.Death),
	}
}

func makeGroupSheetDate(in *gedcom.Event) (out groupSheetDate) {
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

func makeFamiliesAsChild(in viewGroupSheetInputs, target *gedcom.IndividualRecord) ([]familiesAsChild, error) {
	out := make([]familiesAsChild, len(target.FamiliesAsChild))

	for i, famID := range target.FamiliesAsChild {
		fam, ok := in.familiesByID[famID]
		if !ok {
			return nil, fmt.Errorf("family with ID %q not found", in.targetID)
		}

		parents := make([]*groupSheetSimplePerson, len(fam.ParentXrefs))
		for j, parentID := range fam.ParentXrefs {
			individual, ok := in.peopleByID[parentID]
			if !ok {
				return nil, fmt.Errorf("parent with ID %q not found", parentID)
			}
			parents[j] = makeGroupSheetPerson(individual)
			parents[j].Role = "parent"
		}

		siblings := make([]*groupSheetSimplePerson, 0, len(fam.ChildXrefs))
		for _, childID := range fam.ChildXrefs {
			if childID == in.targetID {
				continue
			}

			individual, ok := in.peopleByID[childID]
			if !ok {

				return nil, fmt.Errorf("sibling with ID %q not found", childID)
			}
			siblings = append(siblings, makeGroupSheetPerson(individual))
			siblings[len(siblings)-1].Role = "sibling"
		}

		out[i] = familiesAsChild{
			ID:       famID,
			Parents:  parents,
			Siblings: slices.Clip(siblings),
		}
	}

	return out, nil
}

func makeFamiliesAsPartner(in viewGroupSheetInputs, target *gedcom.IndividualRecord) ([]familiesAsPartner, error) {
	out := make([]familiesAsPartner, len(target.FamiliesAsPartner))

	for i, famID := range target.FamiliesAsPartner {
		fam, ok := in.familiesByID[famID]
		if !ok {
			return nil, fmt.Errorf("family with ID %q not found", in.targetID)
		}

		var partner *groupSheetSimplePerson
		for _, parentID := range fam.ParentXrefs {
			if parentID == in.targetID {
				continue
			}

			individual, ok := in.peopleByID[parentID]
			if !ok {
				return nil, fmt.Errorf("partner with ID %q not found", parentID)
			}

			partner = makeGroupSheetPerson(individual)
			partner.Role = "partner"
		}

		children := make([]*groupSheetSimplePerson, len(fam.ChildXrefs))
		for j, childID := range fam.ChildXrefs {
			individual, ok := in.peopleByID[childID]
			if !ok {
				return nil, fmt.Errorf("child with ID %q not found", childID)
			}
			children[j] = makeGroupSheetPerson(individual)
			children[j].Role = "child"
		}

		out[i] = familiesAsPartner{
			ID:       famID,
			Partner:  partner,
			Children: children,
		}
	}

	return out, nil
}

func formatGroupSheetView(w io.Writer, in *groupSheetView) error {
	var personView, familiesAsChild, familiesAsPartner strings.Builder
	{
		person := table.New().
			Headers("id", "name", "birth_date", "birth_place", "death_date", "death_place").
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					return styleTableHeader
				}
				return styleTableRow
			}).
			Row(in.Person.ID, in.Person.Name, in.Person.Birth.Date, in.Person.Birth.Place, in.Person.Death.Date, in.Person.Death.Place)

		personView.WriteString(styleBold.Render("person") + "\n")
		personView.WriteString(person.Render() + "\n")
		for _, note := range in.Notes {
			personView.WriteString(styleFaint.Width(80).Render(note) + "\n")
		}
	}

	familyHeaders := []string{"role", "name", "birth_date", "birth_place", "death_date", "death_place"}
	{
		familyTable := table.New().
			Headers(familyHeaders...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					return styleTableHeader
				}
				return styleTableRow
			})
		for _, fam := range in.FamiliesAsChild {
			for _, p := range fam.Parents {
				familyTable = familyTable.Row(p.Role, p.Name, p.Birth.Date, p.Birth.Place, p.Death.Date, p.Death.Place)
			}
			for _, p := range fam.Siblings {
				familyTable = familyTable.Row(p.Role, p.Name, p.Birth.Date, p.Birth.Place, p.Death.Date, p.Death.Place)
			}
		}

		familiesAsChild.WriteString(styleBold.Render("families as child") + "\n")
		familiesAsChild.WriteString(familyTable.Render() + "\n")
	}
	{
		familyTable := table.New().
			Headers(familyHeaders...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					return styleTableHeader
				}
				return styleTableRow
			})
		for _, fam := range in.FamiliesAsPartner {
			if fam.Partner != nil {
				familyTable = familyTable.Row(fam.Partner.Role, fam.Partner.Name, fam.Partner.Birth.Date, fam.Partner.Birth.Place, fam.Partner.Death.Date, fam.Partner.Death.Place)
			}
			for _, p := range fam.Children {
				familyTable = familyTable.Row(p.Role, p.Name, p.Birth.Date, p.Birth.Place, p.Death.Date, p.Death.Place)
			}
		}

		familiesAsPartner.WriteString(styleBold.Render("families as partner") + "\n")
		familiesAsPartner.WriteString(familyTable.Render() + "\n")
	}

	_, err := fmt.Fprintln(w, lipgloss.JoinVertical(lipgloss.Center, personView.String(), familiesAsChild.String(), familiesAsPartner.String()))
	return err
}
