package cmd

import (
	"strings"

	"github.com/rafaelespinoza/ged/internal/entity"
)

const fzfLineFieldSeparator = "\t"

func makeFZFPeople(fieldDelimiter string, people []*entity.Person) (out []string) {
	out = make([]string, len(people))
	var line []string

	for i, person := range people {
		// Set cap to maximum number of non-empty fields you might have (assuming 2 parents).
		line = make([]string, 0, 6)

		line = append(line, person.ID)
		line = append(line, "Name:"+person.Name.Full())
		if d := formatDate(person.Birthdate); d != nil {
			line = append(line, "Birth:"+*d)
		}
		if d := formatDate(person.Deathdate); d != nil {
			line = append(line, "Death:"+*d)
		}
		for _, parent := range person.Parents {
			line = append(line, "Parent:"+parent.Name.Full())
		}

		out[i] = strings.Join(line, fieldDelimiter)
	}

	return
}
