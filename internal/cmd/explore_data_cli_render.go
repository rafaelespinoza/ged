package cmd

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/termenv"
	"github.com/rafaelespinoza/ged/internal/log"
)

// Common style definitions. If you're going use a Style modifier method (ie:
// Width(), Margin()), then be sure to first call Copy(). Otherwise, you may
// notice unintended side effects for other uses of these base styles.
var (
	styleBox           = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).Padding(0, 2)
	styleBold          = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center)
	styleBoldUnderline = lipgloss.NewStyle().Bold(true).Underline(true)
	styleTableHeader   = lipgloss.NewStyle().Bold(true).Underline(true).Align(lipgloss.Left).Padding(0, 1)
	styleTableRow      = lipgloss.NewStyle().Padding(0, 1)
	styleFaint         = lipgloss.NewStyle().Faint(true)
)

func init() {
	// Ensure retention of basic styling; even when the output isn't directly
	// connected to a terminal.
	lipgloss.SetColorProfile(termenv.ANSI)
}

func renderGroupSheetView(w io.Writer, in *groupSheetView) error {
	headerStyles := styleBoldUnderline.Copy().MarginBottom(1)
	var personView, familiesAsChild, familiesAsPartner, events strings.Builder
	{
		personView.WriteString(headerStyles.Render("person") + "\n")
		personView.WriteString(tableizeGroupSheetPeople([]string{"id", "name", "birth_date", "birth_place", "death_date", "death_place"}, in.Person) + "\n")
		for _, note := range in.Notes {
			personView.WriteString(styleFaint.Copy().Width(80).Render(note) + "\n")
		}
	}

	familiesAsChild.WriteString(headerStyles.Render("families as child") + "\n")
	for _, fam := range in.FamiliesAsChild {
		familiesAsChild.WriteString(buildFamilyComponent(fam) + "\n")
	}
	familiesAsPartner.WriteString(headerStyles.Render("families as partner") + "\n")
	for _, fam := range in.FamiliesAsPartner {
		familiesAsPartner.WriteString(buildFamilyComponent(fam) + "\n")
	}

	events.WriteString(headerStyles.Render("events") + "\n")
	if len(in.Events) > 0 {
		events.WriteString(listEvents(in.Events))
	}

	_, err := fmt.Fprintln(w, lipgloss.JoinVertical(lipgloss.Center, personView.String(), familiesAsChild.String(), familiesAsPartner.String(), events.String()))
	return err
}

func buildFamilyComponent(fam groupSheetFamily) string {
	parts := make([]string, 0, 4)
	// unlike most other string building functionality here, this func doesn't
	// need to manually add the \n at the end of each part because that's
	// already taken care of by func lipgloss.JoinVertical

	if fam.Title != "" {
		parts = append(parts, fam.Title)
	}
	if fam.MarriedAt.Date != "" {
		parts = append(parts, styleFaint.Render("parents married on: "+fam.MarriedAt.Date))
	}
	if fam.DivorcedAt.Date != "" {
		parts = append(parts, styleFaint.Render("parents divorced on: "+fam.DivorcedAt.Date))
	}
	people := append(fam.Parents, fam.Children...)
	if len(people) > 0 {
		columns := []string{"role", "name", "birth_date", "birth_place", "death_date", "death_place"}
		parts = append(parts, tableizeGroupSheetPeople(columns, people...))
	}

	return lipgloss.JoinVertical(lipgloss.Center, slices.Clip(parts)...)
}

func renderMutualRelationship(w io.Writer, in mutualRelationship) error {
	var relationships, commonEntities strings.Builder

	{
		// Build a card-like component for each relationship compared and put
		// them side-by-side.
		d1 := in.Person1.Name + " is the " + styleBold.Render(in.Relationship1.Description) + " of " + in.Person2.Name
		d2 := in.Person2.Name + " is the " + styleBold.Render(in.Relationship2.Description) + " of " + in.Person1.Name

		p1 := buildPersonVertically(in.Person1)
		p2 := buildPersonVertically(in.Person2)

		r1 := buildRelationshipComponent(d1, p1, p2, in.Relationship1)
		r2 := buildRelationshipComponent(d2, p2, p1, in.Relationship2)

		relationships.WriteString(styleBold.Render("relationships") + "\n")
		relationships.WriteString(
			styleFaint.Render(`How is person 1 related to person 2?
Describe the relationship, and enumerate the path to a common ancestor or union.
Also show the inverse: from person 2 to person 1.`) + "\n",
		)
		relationships.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, styleBox.Render(r1), styleBox.Render(r2)) + "\n")
	}

	{
		// Build a view of the common ancestor or union. Only display if
		// applicable to the relationships.
		common := make([]string, 0)
		columns := []string{"name", "birth_date", "birth_place", "death_date", "death_place"}
		if in.CommonPerson != nil {
			var b strings.Builder
			b.WriteString(styleBold.Render("common ancestor") + "\n")
			b.WriteString(tableizeGroupSheetPeople(columns, in.CommonPerson) + "\n")
			common = append(common, b.String())
		}
		if len(in.Union) > 0 {
			var b strings.Builder
			b.WriteString(styleBold.Render("union") + "\n")
			b.WriteString(tableizeGroupSheetPeople(columns, in.Union...) + "\n")
			common = append(common, b.String())
		}

		commonEntities.WriteString(
			styleBold.Render("common entities") + "\n",
		)
		commonEntities.WriteString(
			styleFaint.Render(`If the people are related by blood, who is their common ancestor?
If related by law, through which union?`) + "\n",
		)
		commonEntities.WriteString(
			lipgloss.JoinHorizontal(lipgloss.Top, common...) + "\n",
		)
	}

	_, err := fmt.Fprintln(w, lipgloss.JoinVertical(lipgloss.Center, relationships.String(), commonEntities.String()))
	return err
}

// buildPersonVertically formats the person fields in a vertical orientation. It
// ensures that the field names and values are aligned in a tabular fashion.
func buildPersonVertically(in *groupSheetSimplePerson) string {
	headerStyles := styleBoldUnderline.Copy().MarginRight(2)
	var columnNames, columnValues strings.Builder
	columns := []struct{ Key, Val string }{
		{"name", in.Name},
		{"birth_date", in.Birth.Date},
		{"birth_place", in.Birth.Place},
		{"death_date", in.Death.Date},
		{"death_place", in.Death.Place},
		{"id", in.ID},
	}
	tail := "\n"
	for i, col := range columns {
		if i == len(columns)-1 {
			tail = ""
		}
		columnNames.WriteString(headerStyles.Render(col.Key) + tail)
		columnValues.WriteString(col.Val + tail)
	}

	out := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, columnNames.String()),
		lipgloss.JoinVertical(lipgloss.Left, columnValues.String()),
	)
	return styleBox.Copy().BorderForeground(lipgloss.ANSIColor(termenv.ANSIBrightBlack)).Render(out)
}

func buildRelationshipComponent(desc, p1, p2 string, rel *relationship) string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		desc,
		lipgloss.JoinHorizontal(lipgloss.Top, p1, p2),
		"path to common entity",
		tableizeGroupSheetPeople([]string{"name", "birth_date", "birth_place", "death_date", "death_place"}, rel.Path...),
	)
}

func tableizeGroupSheetPeople(columnNames []string, people ...*groupSheetSimplePerson) string {
	table := table.New().
		Headers(columnNames...).
		StyleFunc(getTableRowStyle).
		BorderRow(true).
		BorderStyle(styleFaint)
	for _, p := range people {
		values := make([]string, len(columnNames))
		var value string
		for j, column := range columnNames {
			switch column {
			case "id":
				value = p.ID
			case "role":
				value = p.Role
			case "name":
				value = p.Name
			case "birth_date":
				value = p.Birth.Date
			case "birth_place":
				value = p.Birth.Place
			case "death_date":
				value = p.Death.Date
			case "death_place":
				value = p.Death.Place
			default:
				log.Warn(context.TODO(), map[string]any{"column_name": column}, "tableizeGroupSheetPeople: unmapped column")
			}
			values[j] = value
		}
		table = table.Row(values...)
	}

	return table.Render()
}

func listEvents(in []*groupSheetEvent) string {
	out := table.New().
		Headers("date", "place", "type", "notes").
		StyleFunc(getTableRowStyle).
		BorderRow(true).
		BorderStyle(styleFaint)

	wrappingStyle := styleTableRow.Copy().Width(40)

	for _, ev := range in {
		out = out.Row(
			ev.Date.Date,
			wrappingStyle.Render(ev.Date.Place),
			ev.Type,
			wrappingStyle.Render(strings.Join(ev.Notes, "\n")),
		)
	}
	return out.Render()
}

func getTableRowStyle(row, col int) lipgloss.Style {
	if row == 0 {
		return styleTableHeader
	}

	return styleTableRow
}
