package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	styleBox           = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Padding(0, 2)
	styleBold          = lipgloss.NewStyle().Bold(true)
	styleBoldUnderline = lipgloss.NewStyle().Underline(true)
	styleTableHeader   = lipgloss.NewStyle().Underline(true).Align(lipgloss.Left).Padding(0, 1)
	styleTableRow      = lipgloss.NewStyle().Padding(0, 1)
	styleFaint         = lipgloss.NewStyle().Align(lipgloss.Left).Faint(true)
)

func formatMutualRelationship(w io.Writer, in mutualRelationship) error {
	var relationships, commonEntities strings.Builder

	{
		// Build a card-like component for each relationship compared and put
		// them side-by-side.
		p1 := buildPersonComponent(in.Person1)
		p2 := buildPersonComponent(in.Person2)
		r1 := buildRelationshipComponent("relationship_1: from person_1 to person_2", p1, p2, in.Relationship1)
		r2 := buildRelationshipComponent("relationship_2: from person_2 to person_1", p2, p1, in.Relationship2)

		relationships.WriteString(
			styleBold.Align(lipgloss.Center).Render("relationships") + "\n",
		)
		relationships.WriteString(
			styleFaint.Render(`How is person 1 related to person 2?
Describe the relationship, and enumerate the path to a common ancestor or union.
Also show the inverse: from person 2 to person 1.`) + "\n",
		)
		relationships.WriteString(
			lipgloss.JoinHorizontal(lipgloss.Top, styleBox.Render(r1), styleBox.Render(r2)) + "\n",
		)
	}

	{
		// Build a view of the common ancestor or union. Only display if
		// applicable to the relationships.
		common := make([]string, 0)
		if in.CommonPerson != nil {
			commonPerson := table.New().
				Headers("name", "birth_date", "birth_place", "death_date", "death_place").
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == 0 {
						return styleTableHeader
					}
					return styleTableRow
				}).
				Row(in.CommonPerson.Name, in.CommonPerson.Birth.Date, in.CommonPerson.Birth.Place, in.CommonPerson.Death.Date, in.CommonPerson.Death.Place)

			var b strings.Builder
			b.WriteString(styleBold.Align(lipgloss.Center).Render("common ancestor") + "\n")
			b.WriteString(commonPerson.Render() + "\n")
			common = append(common, b.String())
		}
		if len(in.Union) > 0 {
			union := table.New().
				Headers("name", "birth_date", "birth_place", "death_date", "death_place").
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == 0 {
						return styleTableHeader
					}
					return styleTableRow
				})
			for _, person := range in.Union {
				union = union.Row(person.Name, person.Birth.Date, person.Birth.Place, person.Death.Date, person.Death.Place)
			}

			var b strings.Builder
			b.WriteString(styleBold.Align(lipgloss.Center).Render("union") + "\n")
			b.WriteString(union.Render() + "\n")
			common = append(common, b.String())
		}

		commonEntities.WriteString(
			styleBold.Align(lipgloss.Center).Render("common entities") + "\n",
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

func buildPersonComponent(in *groupSheetSimplePerson) string {
	// format/display person fields in vertical orientation, ensure the keys and
	// values are aligned.
	headerStyles := styleBoldUnderline.MarginRight(2)
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
	return styleBox.Render(out)
}

func buildRelationshipComponent(title, p1, p2 string, rel *relationship) string {
	var b strings.Builder

	b.WriteString(styleBold.MarginBottom(1).Render(title) + "\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, p1, p2) + "\n")
	b.WriteString(fmt.Sprintf("%s, generations_removed: %d", rel.Description, rel.GenerationsRemoved) + "\n")

	path := table.New().
		Headers("name", "birth_date", "birth_place", "death_date", "death_place").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return styleBoldUnderline.Align(lipgloss.Left).Padding(0, 1)
			}
			return styleTableRow
		})
	for _, person := range rel.Path {
		path = path.Row(person.Name, person.Birth.Date, person.Birth.Place, person.Death.Date, person.Death.Place)
	}
	b.WriteString(path.Render())

	return b.String()
}
