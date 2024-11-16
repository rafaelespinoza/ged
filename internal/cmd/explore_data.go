package cmd

import (
	"fmt"
	"strings"

	"github.com/rafaelespinoza/alf"
)

func makeExploreData(name string) alf.Directive {
	out := alf.Delegator{
		Description: "view GEDCOM data",
		Subs: map[string]alf.Directive{
			"show":   makeExploreDataShow(name, "show"),
			"relate": makeExploreDataRelate(name, "relate"),
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
	groupSheetView struct {
		Person            *groupSheetSimplePerson
		Notes             []string
		FamiliesAsChild   []groupSheetFamily
		FamiliesAsPartner []groupSheetFamily
		Events            []*groupSheetEvent
	}
	groupSheetSimplePerson struct {
		ID    string
		Role  string
		Name  string
		Birth groupSheetDate
		Death groupSheetDate
	}
	groupSheetFamily struct {
		ID         string
		Title      string
		MarriedAt  groupSheetDate
		DivorcedAt groupSheetDate
		Parents    []*groupSheetSimplePerson
		Children   []*groupSheetSimplePerson
	}
	groupSheetEvent struct {
		Date  groupSheetDate
		Type  string
		Notes []string
	}
	groupSheetDate struct {
		Date  string
		Place string
	}
)
