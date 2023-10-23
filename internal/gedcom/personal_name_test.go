package gedcom

import (
	"context"
	"testing"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"
)

func TestPersonalName(t *testing.T) {
	type Testcase struct {
		Name     string // the name of the test case
		Line     *gedcom7.Line
		Sublines []*gedcom7.Line
		Input    string // the input name to the function we're testing
		Exp      PersonalName
	}

	runTest := func(t *testing.T, test Testcase) {
		input := gedcom.NewNode(nil, test.Line)
		for _, subline := range test.Sublines {
			input = input.AddSubnode(subline)
		}
		got, err := parsePersonalName(context.Background(), test.Line, input.GetSubnodes())
		if err != nil {
			t.Fatal(err)
		}

		if got.Given != test.Exp.Given {
			t.Errorf("wrong Given, got %q, exp %q", got.Given, test.Exp.Given)
		}
		if got.Surname != test.Exp.Surname {
			t.Errorf("wrong Surname, got %q, exp %q", got.Surname, test.Exp.Surname)
		}
		if got.NameSuffix != test.Exp.NameSuffix {
			t.Errorf("wrong NameSuffix, got %q, exp %q", got.NameSuffix, test.Exp.NameSuffix)
		}
	}

	t.Run("payload only", func(t *testing.T) {
		tests := []Testcase{
			{
				Name: "givenname surname",
				Line: &gedcom7.Line{Payload: "Santa /Clause/"},
				Exp:  PersonalName{Given: "Santa", Surname: "Clause"},
			},
			{
				Name: "multiple part surname",
				Line: &gedcom7.Line{Payload: "Claus /van Rosenvelt/"},
				Exp:  PersonalName{Given: "Claus", Surname: "van Rosenvelt"},
			},
			{
				Name: "multiple part givenname",
				Line: &gedcom7.Line{Payload: "John Fitzgerald /Kennedy/"},
				Exp:  PersonalName{Given: "John Fitzgerald", Surname: "Kennedy"},
			},
			{
				Name: "name suffix",
				Line: &gedcom7.Line{Payload: "Sammy George /Davis/ Jr."},
				Exp:  PersonalName{Given: "Sammy George", Surname: "Davis", NameSuffix: "Jr."},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
		}
	})

	t.Run("name pieces", func(t *testing.T) {
		tests := []Testcase{
			{
				Name:     "givenname surname",
				Line:     &gedcom7.Line{Payload: "Benjamin /Franklin/"},
				Sublines: []*gedcom7.Line{{Tag: "GIVN", Payload: "Benjamin"}, {Tag: "SURN", Payload: "Franklin"}},
				Exp:      PersonalName{Given: "Benjamin", Surname: "Franklin"},
			},
			{
				Name:     "multiple part surname",
				Line:     &gedcom7.Line{Payload: "Alice /de Bunbury/"},
				Sublines: []*gedcom7.Line{{Tag: "GIVN", Payload: "Alice"}, {Tag: "SURN", Payload: "de Bunbury"}},
				Exp:      PersonalName{Given: "Alice", Surname: "de Bunbury"},
			},
			{
				Name:     "multiple part givenname",
				Line:     &gedcom7.Line{Payload: "Robert Francis /Kennedy/"},
				Sublines: []*gedcom7.Line{{Tag: "GIVN", Payload: "Robert Francis"}, {Tag: "SURN", Payload: "Kennedy"}},
				Exp:      PersonalName{Given: "Robert Francis", Surname: "Kennedy"},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
		}
	})
}
