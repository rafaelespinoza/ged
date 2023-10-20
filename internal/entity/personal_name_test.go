package entity

import "testing"

func TestPersonalNameFull(t *testing.T) {
	tests := []struct {
		Name string
		In   PersonalName
		Exp  string
	}{
		{
			Name: "all parts",
			In: PersonalName{
				Forename:     "First",
				Middle:       "Middle",
				Surname:      "Surname",
				BirthSurname: "BirthSurname",
				Nickname:     "Nick",
				Suffix:       "OBE",
			},
			Exp: `First "Nick" Middle BirthSurname Surname OBE`,
		},
		{
			Name: "some parts",
			In: PersonalName{
				Forename: "First",
				Middle:   "Middle",
				Surname:  "Surname",
			},
			Exp: `First Middle Surname`,
		},
		{
			Name: "no parts",
			In:   PersonalName{},
			Exp:  ``,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := test.In.Full()

			if got != test.Exp {
				t.Errorf("got %q, exp %q", got, test.Exp)
			}
		})
	}
}
