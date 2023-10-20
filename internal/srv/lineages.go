package srv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rafaelespinoza/ged/internal/entity"
)

// makeLineages computes the familial relationship between two ancestral
// lines. Inputs line1 and line2 are both ancestral lines that should have the
// IDs of each person as the first item. It's based off of the PHP-based sample
// in https://stackoverflow.com/a/1087345.
func makeLineages(r *relator, line1, line2 []string) (out1, out2 entity.Lineage, err error) {
	out1 = entity.Lineage{
		GenerationsRemoved: len(line1) - len(line2),
		CommonAncestors:    buildCommonAncestors(r, line1),
	}
	out2 = entity.Lineage{
		GenerationsRemoved: len(line2) - len(line1),
		CommonAncestors:    buildCommonAncestors(r, line2),
	}

	// calculate the number of generations from the start of each ancestral line
	// to the end. Subtract 1 because the 1st item includes the ID of the person
	// from the initial request.
	dist1 := len(line1) - 1
	dist2 := len(line2) - 1

	if dist1 == 0 && dist2 == 0 && line1[0] == line2[0] {
		out1.Type = entity.Self
		out2.Type = entity.Self
	} else if dist1 == 0 { // direct descendant - parent
		out1.Type = entity.Parent
		out2.Type = entity.Child
	} else if dist2 == 0 { // direct descendant - child
		out1.Type = entity.Child
		out2.Type = entity.Parent
	} else if dist1 == dist2 { // equal distance - either siblings or cousins
		if dist1 == 1 {
			out1.Type = entity.Sibling
			out2.Type = entity.Sibling
		} else {
			out1.Type = entity.Cousin
			out2.Type = entity.Cousin
			dist1 *= -1
			dist2 *= -1
		}
	} else if dist1 == 1 {
		out1.Type = entity.AuntUncle
		out2.Type = entity.NieceNephew
	} else if dist2 == 1 {
		out1.Type = entity.NieceNephew
		out2.Type = entity.AuntUncle
	} else { // cousins, generationally removed
		out1.Type = entity.Cousin
		out2.Type = entity.Cousin
		dist1 *= -1
		dist2 *= -1
	}

	desc, _, err := describeLineage(out1.Type, out1.GenerationsRemoved, dist1)
	if err != nil {
		err = fmt.Errorf("lineage 1: %w", err)
		return
	}
	out1.Description = desc

	desc, _, err = describeLineage(out2.Type, out2.GenerationsRemoved, dist2)
	if err != nil {
		err = fmt.Errorf("lineage 2: %w", err)
		return
	}
	out2.Description = desc

	return
}

func buildCommonAncestors(r *relator, path []string) []entity.Person {
	out := make([]entity.Person, len(path))

	for i, personID := range path {
		person, _ := r.lookupOne(personID)
		out[i] = entity.Person{
			ID:        person.ID,
			Name:      person.Name,
			Birthdate: person.Birthdate,
			Deathdate: person.Deathdate,
		}
	}

	return out
}

func describeLineage(t entity.LineageType, generationsRemoved int, generationsSinceCommonAncestor int) (out string, related bool, err error) {
	related = true

	switch t {
	case entity.Child, entity.NieceNephew:
		if generationsRemoved <= 0 {
			err = fmt.Errorf("generations removed (%d) must be > 0 for child", generationsRemoved)
		} else {
			switch generationsRemoved {
			case 1:
				out = t.String()
			case 2:
				out = "grand " + t.String()
			default:
				out = strings.Repeat("great ", generationsRemoved-2) + "grand " + t.String()
			}
		}
	case entity.Self, entity.Sibling:
		out = t.String()
	case entity.Parent:
		if generationsRemoved >= 0 {
			err = fmt.Errorf("generations removed (%d) must be < 0 for parent", generationsRemoved)
		} else {
			switch generationsRemoved {
			case -1:
				out = t.String()
			case -2:
				out = "grand " + t.String()
			default:
				out = strings.Repeat("great ", -generationsRemoved-2) + "grand " + t.String()
			}
		}
	case entity.AuntUncle:
		if generationsRemoved >= 0 {
			err = fmt.Errorf("generations removed (%d) must be < 0 for auntuncle", generationsRemoved)
		} else {
			switch generationsRemoved {
			case -1:
				out = t.String()
			case -2:
				out = "great " + t.String()
			default:
				out = strings.Repeat("great ", -generationsRemoved-2) + "grand " + t.String()
			}
		}
	case entity.Cousin:
		// to be cousins, the most recent common ancestor must be from 2 or more
		// generations in the past.
		if generationsSinceCommonAncestor > -2 {
			// Lineage seems to actually be parent-child or sibling; or it violates causality.
			err = fmt.Errorf("generations since common ancestor (%d) must be < -1", generationsSinceCommonAncestor)
		} else {
			// TODO: clean up this mapping for generationally removed cousins
			// from the younger cousin's perspective. Not sure why it works so
			// far, but it does :shrug:
			n := -generationsSinceCommonAncestor - 1

			if generationsRemoved == 0 {
				out = makeOrdinalSuffix(n) + " " + t.String()
			} else if generationsRemoved >= -20 && generationsRemoved <= 20 {
				if generationsRemoved < 0 {
					generationsRemoved *= -1
				} else {
					n -= generationsRemoved
				}

				ordSuff := makeOrdinalSuffix(n)
				out = ordSuff + " " + t.String() + " " + strconv.Itoa(generationsRemoved) + "x removed"
			} else {
				out = "distant " + t.String()
			}
		}
	default:
		related = false
	}

	if err != nil {
		related = false
	}

	return
}

func makeOrdinalSuffix(n int) (out string) {
	if n%100 >= 11 && n%100 < 14 { // 11th, 12th, 13th
		out = fmt.Sprintf("%dth", n)
		return
	}

	var suffix string
	switch n % 10 {
	case 1:
		suffix = "st"
	case 2:
		suffix = "nd"
	case 3:
		suffix = "rd"
	default:
		suffix = "th"
	}

	out = fmt.Sprintf("%d%s", n, suffix)
	return
}
