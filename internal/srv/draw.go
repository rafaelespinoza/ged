package srv

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/log"

	mermaid_go "github.com/dreampuf/mermaid.go"
)

type MermaidFlowchartParams struct {
	Out       io.Writer
	Direction string
	DisplayID bool
	People    []*entity.Person
	Unions    []*entity.Union
}

// ValidFlowchartDirections defines Mermaid-specific orientations for a
// flowchart. More info is at https://mermaid.js.org/syntax/flowchart.html#direction.
var ValidFlowchartDirections = []string{"TB", "TD", "BT", "RL", "LR"}

func MakeMermaidFlowchart(ctx context.Context, p MermaidFlowchartParams) error {
	if !slices.Contains(ValidFlowchartDirections, p.Direction) {
		return fmt.Errorf("invalid Direction %q, valid ones are: %q", p.Direction, ValidFlowchartDirections)
	}

	tmpl, err := template.New("").Parse(mermaidFlowchartFamilyTree)
	if err != nil {
		return err
	}

	type ExecData struct {
		FlowChartDirection string
		PeopleIDs          []string
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
	// GEDCOM-formatted ID will always start and end with @. So it's OK to have
	// the @ symbol in the node label, just can't put it in the node ID.
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
		var dateSpan string
		if union.StartDate != nil || union.EndDate != nil {
			dateSpan = formatYear(union.StartDate) + " - " + formatYear(union.EndDate)
		}
		unionsByID[union.ID] = &drawUnionOutput{
			ID:        union.ID,
			Person1ID: p1,
			Person2ID: p2,
			DateSpan:  dateSpan,
			ChildIDs:  childIDs,
		}
	}

	for i, person := range p.People {
		var originalID string
		if p.DisplayID {
			// Retain the original ID for display b/c that's the ID used in the Relate people functionality.
			originalID = person.ID
		}

		person.ID = stripAtSign(person.ID)
		allPeopleIDs[i] = person.ID

		var abbreviatedName string
		if person.Name.Forename != "" && person.Name.Surname != "" {
			abbreviatedName = person.Name.Forename[:1] + ". " + person.Name.Surname
		}

		var dateSpan string
		if person.Birthdate != nil || person.Deathdate != nil {
			dateSpan = formatYear(person.Birthdate) + " - " + formatYear(person.Deathdate)
		}

		displayPersonData := drawPersonOutput{
			ID:              person.ID,
			OriginalID:      originalID,
			Fullname:        strings.ReplaceAll(person.Name.Full(), `"`, `#quot;`),
			AbbreviatedName: abbreviatedName,
			DateSpan:        dateSpan,
		}
		peopleByID[person.ID] = &displayPersonData
	}

	return tmpl.Execute(p.Out, ExecData{
		FlowChartDirection: p.Direction,
		PeopleIDs:          allPeopleIDs,
		PeopleByID:         peopleByID,
		UnionsByID:         unionsByID,
	})
}

type drawPersonOutput struct {
	ID              string
	OriginalID      string
	Fullname        string
	AbbreviatedName string
	DateSpan        string
}

type drawUnionOutput struct {
	ID        string
	Person1ID string
	Person2ID string
	DateSpan  string
	ChildIDs  []string
}

const mermaidFlowchartFamilyTree = `flowchart {{$.FlowChartDirection}}

classDef unionNode height:5rem,width:10rem,display:inline-block;

%% define people

{{- range $_, $id := $.PeopleIDs}}
	{{$person := index $.PeopleByID $id}}
	{{$id}}("{{$person.Fullname}}
{{$person.DateSpan}}
{{with $person.OriginalID}}({{.}}){{end -}}
")
{{- end}}

%% define unions

{{range $_, $union := $.UnionsByID}}
	{{- $person1 := index $.PeopleByID $union.Person1ID}}
	{{- $person2 := index $.PeopleByID $union.Person2ID}}

	%% "{{with $person1}}{{.Fullname}}{{end}}" and "{{with $person2}}{{.Fullname}}{{end}}"
	{{$union.ID}}>"
{{with $person1}}{{.AbbreviatedName}}{{else}}unknown{{end}}
+
{{with $person2}}{{.AbbreviatedName}}{{else}}unknown{{end}}
{{with $union.DateSpan}}{{.}}{{else}} {{- end}}{{/* as a fallback, leave empty space so that Mermaid can render */}}
"]:::unionNode

	{{with $union.Person1ID}}{{.}}-...->{{$union.ID}}{{end}}
	{{with $union.Person2ID}}{{.}}-...->{{$union.ID}}{{end}}
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

type MermaidRenderer interface {
	DrawSVG(ctx context.Context, w io.Writer) error
	DrawPNG(ctx context.Context, w io.Writer, scale float64) error
	Close() error
}

func NewMermaidRenderer(ctx context.Context, r io.Reader) (MermaidRenderer, error) {
	const flowchartStatements = `mermaid.initialize({
		maxTextSize: 200000,	// default is 50000
		flowchart: { defaultRenderer: "elk" },
	});`
	re, err := mermaid_go.NewRenderEngine(ctx, flowchartStatements)
	if err != nil {
		return nil, err
	}
	contents, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &mermaidRenderer{re, string(contents)}, nil
}

type mermaidRenderer struct {
	re       *mermaid_go.RenderEngine
	contents string
}

func (m *mermaidRenderer) DrawSVG(ctx context.Context, w io.Writer) error {
	out, err := m.re.Render(m.contents)
	if err != nil {
		return err
	}

	// Need to fix the invalid <br> tags.
	// See issue number 74 on the mermaid-cli github repo at https://github.com/mermaid-js/mermaid-cli
	out = strings.ReplaceAll(out, `<br>`, `<br/>`)

	n, err := w.Write([]byte(out))
	if err != nil {
		return err
	}
	log.Debug(ctx, map[string]any{"num_bytes_written": n, "method": "DrawSVG"}, "wrote mermaid data")
	return nil
}

func (m *mermaidRenderer) DrawPNG(ctx context.Context, w io.Writer, scale float64) error {
	out, box, err := m.re.RenderAsScaledPng(m.contents, scale)
	if err != nil {
		return err
	}
	n, err := w.Write(out)
	if err != nil {
		return err
	}
	log.Debug(ctx, map[string]any{"num_bytes_written": n, "scale": scale, "method": "DrawPNG", "box": box}, "wrote mermaid data")
	return nil
}

func (m *mermaidRenderer) Close() error {
	m.re.Cancel()
	return nil
}
