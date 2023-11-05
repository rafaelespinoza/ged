package srv

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/gedcom"
	"github.com/rafaelespinoza/ged/internal/log"
)

func ParseGedcom(ctx context.Context, r io.Reader) ([]*entity.Person, []*entity.Union, error) {
	records, err := gedcom.ReadRecords(ctx, r)
	if err != nil {
		return nil, nil, err
	}

	log.Info(ctx, map[string]any{"records": records}, "converted gedcom records")

	gedcomFamiliesByID := make(map[string]*gedcom.FamilyRecord, len(records.Families))
	for _, family := range records.Families {
		gedcomFamiliesByID[family.Xref] = family
	}

	peopleByGCID, err := convertGedcomPeople(ctx, records.Individuals, gedcomFamiliesByID)
	if err != nil {
		return nil, nil, err
	}

	unionsByGCID, err := convertGedcomFamilies(ctx, records.Families, peopleByGCID)
	if err != nil {
		return nil, nil, err
	}

	people := make([]*entity.Person, len(records.Individuals))
	var person *entity.Person
	for i, individual := range records.Individuals {
		person = peopleByGCID[individual.Xref]
		people[i] = person
	}

	unions := make([]*entity.Union, len(records.Families))
	var union *entity.Union
	for i, family := range records.Families {
		union = unionsByGCID[family.Xref]
		unions[i] = union
	}

	return people, unions, nil
}

func convertGedcomPeople(ctx context.Context, records []*gedcom.IndividualRecord, gedcomFamiliesByID map[string]*gedcom.FamilyRecord) (map[string]*entity.Person, error) {
	out := make(map[string]*entity.Person, len(records))

	for _, individual := range records {
		var (
			inputName gedcom.PersonalName
			birthdate *entity.Date
			deathdate *entity.Date
			err       error
		)

		if len(individual.Names) > 0 {
			inputName = individual.Names[0]
		}
		if individual.Birth != nil {
			birthdate, err = entity.NewDate(individual.Birth.Date, individual.Birth.DateRange)
			if err != nil {
				log.Error(ctx, map[string]any{"individual": individual}, err, "invalid Birth.Date")
				return nil, err
			}
		}
		if individual.Death != nil {
			deathdate, err = entity.NewDate(individual.Death.Date, individual.Death.DateRange)
			if err != nil {
				log.Error(ctx, map[string]any{"individual": individual}, err, "invalid Death.Date")
				return nil, err
			}
		}

		out[individual.Xref] = &entity.Person{
			ID: individual.Xref,
			Name: entity.PersonalName{
				Forename: inputName.Given,
				Nickname: inputName.Nickname,
				Surname:  inputName.Surname,
				Suffix:   inputName.NameSuffix,
			},
			Birthdate: birthdate,
			Deathdate: deathdate,
		}
	}

	for _, individual := range records {
		parentTuples := make([]*entity.Person, 0, len(individual.FamiliesAsChild)*2)
		for _, famID := range individual.FamiliesAsChild {
			familyRecord, ok := gedcomFamiliesByID[famID]
			if !ok {
				return nil, fmt.Errorf("gedcom family as child %q not found for individual %q", famID, individual.Xref)
			}

			for _, parentID := range familyRecord.ParentXrefs {
				parent, ok := out[parentID]
				if !ok {
					return nil, fmt.Errorf("entity parent %q from family %q not found for individual as child %q", parentID, famID, individual.Xref)
				}
				parentTuples = append(parentTuples, simplifyPerson(parent))
			}
		}

		childTuples := make([]*entity.Person, 0, len(individual.FamiliesAsPartner)*2)
		spouseTuples := make([]*entity.Person, 0, len(individual.FamiliesAsPartner))
		for _, famID := range individual.FamiliesAsPartner {
			familyRecord, ok := gedcomFamiliesByID[famID]
			if !ok {
				return nil, fmt.Errorf("gedcom family as partner %q not found for individual %q", famID, individual.Xref)
			}

			for _, spouseID := range familyRecord.ParentXrefs {
				spouse, ok := out[spouseID]
				if !ok {
					return nil, fmt.Errorf("entity spouse %q from family %q not found for individual as partner %q", spouseID, famID, individual.Xref)
				}
				if spouseID == individual.Xref {
					continue
				}
				spouseTuples = append(spouseTuples, simplifyPerson(spouse))
			}

			for _, childID := range familyRecord.ChildXrefs {
				child, ok := out[childID]
				if !ok {
					return nil, fmt.Errorf("entity child %q from family %q not found for individual as partner %q", childID, famID, individual.Xref)
				}
				childTuples = append(childTuples, simplifyPerson(child))
			}
		}

		person := out[individual.Xref]
		person.Parents = slices.Clip(parentTuples)
		person.Children = slices.Clip(childTuples)
		person.Spouses = slices.Clip(spouseTuples)
		out[individual.Xref] = person
	}

	return out, nil
}

func convertGedcomFamilies(ctx context.Context, records []*gedcom.FamilyRecord, peopleByGCID map[string]*entity.Person) (map[string]*entity.Union, error) {
	out := make(map[string]*entity.Union, len(records))

	var err error

	for i, family := range records {
		union := entity.Union{ID: family.Xref}
		if family.MarriedAt != nil {
			union.StartDate, err = entity.NewDate(family.MarriedAt.Date, family.MarriedAt.DateRange)
			if err != nil {
				log.Error(ctx, map[string]any{"family": family}, err, "invalid StartDate")
				return nil, err
			}
		}
		if family.DivorcedAt != nil && family.DivorcedAt.Date != nil {
			union.EndDate, err = entity.NewDate(family.DivorcedAt.Date, family.DivorcedAt.DateRange)
			if err != nil {
				log.Error(ctx, map[string]any{"family": family}, err, "invalid EndDate")
				return nil, err
			}
		}
		if family.AnnulledAt != nil && family.AnnulledAt.Date != nil {
			union.EndDate, err = entity.NewDate(family.AnnulledAt.Date, family.AnnulledAt.DateRange)
			if err != nil {
				log.Error(ctx, map[string]any{"family": family}, err, "invalid EndDate")
				return nil, err
			}
		}

		partners := make([]*entity.Person, len(family.ParentXrefs))
		for i, partnerID := range family.ParentXrefs {
			partner, ok := peopleByGCID[partnerID]
			if !ok {
				return nil, fmt.Errorf("partner %q not found for family %q", partnerID, family.Xref)
			}
			partners[i] = simplifyPerson(partner)
		}

		switch len(partners) {
		case 0:
			return nil, fmt.Errorf("no partner references for family %q", family.Xref)
		case 1:
			union.Person1 = partners[0]
		case 2:
			union.Person1 = partners[0]
			union.Person2 = partners[1]
		default:
			union.Person1 = partners[0]
			union.Person2 = partners[1]

			log.Warn(ctx, map[string]any{
				"xref":        family.Xref,
				"i":           i,
				"partner_ids": family.ParentXrefs[2:],
			}, "discarding extra partner references in family")
		}

		children := make([]*entity.Person, len(family.ChildXrefs))
		for i, childID := range family.ChildXrefs {
			child, ok := peopleByGCID[childID]
			if !ok {
				return nil, fmt.Errorf("child %q not found for family %q", childID, family.Xref)
			}
			children[i] = simplifyPerson(child)
		}
		union.Children = children

		out[family.Xref] = &union
	}

	return out, nil
}

// simplifyPerson intentionally does not copy the Children, Parent, or Spouses
// fields to help keep each output item succinct. This is most beneficial when
// marshaling the results. Without such a limit, you could end up with
// generations upon generations of deeply-nested structures.
func simplifyPerson(in *entity.Person) *entity.Person {
	return &entity.Person{
		ID:        in.ID,
		Name:      in.Name,
		Birthdate: in.Birthdate,
		Deathdate: in.Deathdate,
	}
}
