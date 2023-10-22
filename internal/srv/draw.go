package srv

import (
	"context"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/rafaelespinoza/ged/internal/entity"
)

type DrawParams struct {
	Out    io.Writer
	People []*entity.Person
	Unions []*entity.Union
}

func Draw(ctx context.Context, p DrawParams) error {
	tmpl, err := template.New("").Parse(mermaidFlowchartFamilyTree)
	if err != nil {
		return err
	}

	type ExecData struct {
		PeopleIDs []string
		// PeopleByID is a pointer type to simplify emptiness checks
		// inside of template.
		PeopleByID map[string]*drawPersonOutput
		UnionsByID map[string]*drawUnionOutput
	}

	allPeopleIDs := make([]string, len(p.People))
	peopleByID := make(map[string]*drawPersonOutput, len(p.People))
	unionsByID := make(map[string]*drawUnionOutput)

	// Use this function to prepare ID values for Mermaid diagrams. In
	// particular, for declaring nodes by ID and to link nodes between nodes.
	// Mermaid does not like the @ character in node IDs. However, any
	// GEDCOM-formatted ID will always start and end with @. Should only need to
	// use this so that Mermaid can draw the graph.
	stripAtSign := func(in string) string { return strings.ReplaceAll(in, "@", "") }

	for _, union := range p.Unions {
		var p1, p2 string
		if union.Person1 != nil {
			p1 = stripAtSign(union.Person1.ID)
		}
		if union.Person2 != nil {
			p2 = stripAtSign(union.Person2.ID)
		}

		union.ID = stripAtSign(union.ID)
		childIDs := make([]string, len(union.Children))
		for i, child := range union.Children {
			childIDs[i] = stripAtSign(child.ID)
		}
		unionsByID[union.ID] = &drawUnionOutput{
			ID:        union.ID,
			Person1ID: p1,
			Person2ID: p2,
			StartDate: formatYear(union.StartDate),
			EndDate:   formatYear(union.EndDate),
			ChildIDs:  childIDs,
		}
	}

	for i, person := range p.People {
		// Retain the original ID for display b/c that's the ID used in the Relate people functionality.
		originalID := person.ID

		person.ID = stripAtSign(person.ID)
		allPeopleIDs[i] = person.ID

		peopleByID[person.ID] = &drawPersonOutput{
			ID:         person.ID,
			OriginalID: originalID,
			Fullname:   strings.ReplaceAll(person.Name.Full(), `"`, `#quot;`),
			BirthYear:  formatYear(person.Birthdate),
			DeathYear:  formatYear(person.Deathdate),
		}
	}

	return tmpl.Execute(p.Out, ExecData{PeopleIDs: allPeopleIDs, PeopleByID: peopleByID, UnionsByID: unionsByID})
}

type drawPersonOutput struct {
	ID         string
	OriginalID string
	Fullname   string
	BirthYear  string
	DeathYear  string
}

type drawUnionOutput struct {
	ID        string
	Person1ID string
	Person2ID string
	StartDate string
	EndDate   string
	ChildIDs  []string
}

const mermaidFlowchartFamilyTree = `flowchart TD
%% define people

{{range $_, $id := $.PeopleIDs}}
	{{$person := index $.PeopleByID $id}}
	{{- $id}}("
({{$person.OriginalID}})
{{$person.Fullname}}
{{with $person.BirthYear}}b. {{.}}{{end}} {{with $person.DeathYear}}d. {{.}}{{end}}
")
{{- end}}

%% define unions

{{range $_, $union := $.UnionsByID}}
	{{- $person1 := index $.PeopleByID $union.Person1ID}}
	{{- $person2 := index $.PeopleByID $union.Person2ID}}

	%% "{{with $person1}}{{.Fullname}}{{end}}" and "{{with $person2}}{{.Fullname}}{{end}}"
	{{$union.ID }}[["{{with $union.StartDate}}{{.}}{{end}} - {{with $union.EndDate}}{{.}}{{end}}"]]
	{{$union.Person1ID }}-...->{{$union.ID}}
	{{$union.Person2ID }}-...->{{$union.ID}}
	{{range $_, $childID := $union.ChildIDs}}
	{{$union.ID}} =====> {{$childID}}
	{{- end}}
{{- end}}
`

func formatYear(in *time.Time) string {
	if in == nil {
		return ""
	}
	return in.Format("2006")
}
