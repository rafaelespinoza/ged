package gedcom_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/rafaelespinoza/ged/internal/gedcom"
)

func TestPersonalName(t *testing.T) {
	type Testcase struct {
		Name          string // the name of the test case
		RawData       string // GEDCOM-formatted data for just 1 IndividualRecord
		ExpectedNames []gedcom.PersonalName
	}

	runTest := func(t *testing.T, test Testcase) {
		records, err := gedcom.ReadRecords(context.Background(), strings.NewReader(test.RawData))
		if err != nil {
			t.Fatal(err)
		}

		if len(records.Individuals) != 1 {
			t.Fatalf("expected just 1 individual record; got %d", len(records.Individuals))
		}

		individual := records.Individuals[0]

		if len(individual.Names) != len(test.ExpectedNames) {
			t.Fatalf("wrong number of Names, got %d, exp %d", len(individual.Names), len(test.ExpectedNames))
		}

		for i, got := range individual.Names {
			errMsgPrefix := fmt.Sprintf("item[%d]", i)
			cmpPersonalName(t, errMsgPrefix, got, test.ExpectedNames[i])
		}
	}

	t.Run("payload only", func(t *testing.T) {
		tests := []Testcase{
			{
				Name: "givenname surname",
				RawData: `0 HEAD,
0 @I1@ INDI
1 NAME Santa /Clause/
0 TRLR`,
				ExpectedNames: []gedcom.PersonalName{{Payload: "Santa /Clause/", Given: "Santa", Surname: "Clause"}},
			},
			{
				Name: "multiple part surname",
				RawData: `0 HEAD,
0 @I1@ INDI
1 NAME Claus /van Rosenvelt/
0 TRLR`,
				ExpectedNames: []gedcom.PersonalName{{Payload: "Claus /van Rosenvelt/", Given: "Claus", Surname: "van Rosenvelt"}},
			},
			{
				Name: "multiple part givenname",
				RawData: `0 HEAD,
0 @I1@ INDI
1 NAME John Fitzgerald /Kennedy/
0 TRLR`,
				ExpectedNames: []gedcom.PersonalName{{Payload: "John Fitzgerald /Kennedy/", Given: "John Fitzgerald", Surname: "Kennedy"}},
			},
			{
				Name: "name suffix",
				RawData: `0 HEAD,
0 @I1@ INDI
1 NAME Sammy George /Davis/ Jr.
0 TRLR`,

				ExpectedNames: []gedcom.PersonalName{{Payload: "Sammy George /Davis/ Jr.", Given: "Sammy George", Surname: "Davis", NameSuffix: "Jr."}},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
		}
	})

	t.Run("name pieces", func(t *testing.T) {
		tests := []Testcase{
			{
				Name: "givenname surname",
				RawData: `0 HEAD,
0 @I1@ INDI
1 NAME Benjamin /Franklin/
2 GIVN Benjamin
2 SURN Franklin
0 TRLR`,
				ExpectedNames: []gedcom.PersonalName{{Payload: "Benjamin /Franklin/", Given: "Benjamin", Surname: "Franklin"}},
			},
			{
				Name: "multiple part surname",
				RawData: `0 HEAD,
0 @I1@ INDI
1 NAME Alice /de Bunbury/
2 GIVN Alice
2 SURN de Bunbury
0 TRLR`,
				ExpectedNames: []gedcom.PersonalName{{Payload: "Alice /de Bunbury/", Given: "Alice", Surname: "de Bunbury"}},
			},
			{
				Name: "multiple part givenname",
				RawData: `0 HEAD,
0 @I1@ INDI
1 NAME Robert Francis /Kennedy/
2 GIVN Robert Francis
2 SURN Kennedy
0 TRLR`,
				ExpectedNames: []gedcom.PersonalName{{Payload: "Robert Francis /Kennedy/", Given: "Robert Francis", Surname: "Kennedy"}},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
		}
	})

	t.Run("source citations", func(t *testing.T) {
		runTest(t, Testcase{
			RawData: `0 HEAD,
0 @I1@ INDI
1 NAME Santa /Clause/
2 SOUR @S123@
3 PAGE Foo: 1
2 SOUR @S234@
3 DATA
4 WWW https://example.test/foo/1
0 TRLR`,
			ExpectedNames: []gedcom.PersonalName{
				{
					Payload: "Santa /Clause/",
					Given:   "Santa",
					Surname: "Clause",
					SourceCitations: []*gedcom.SourceCitation{
						{
							Xref: "@S123@",
							Page: "Foo: 1",
						},
						{
							Xref: "@S234@",
							Data: map[string]string{
								"WWW": "https://example.test/foo/1",
							},
						},
					},
				},
			},
		})
	})

	t.Run("multiple names", func(t *testing.T) {
		runTest(t, Testcase{
			RawData: `0 HEAD,
0 @I1@ INDI
1 NAME Santa /Clause/
2 TYPE PROFESSIONAL
2 SOUR @S123@
3 PAGE Foo: 1
1 NAME Kristopher /Kringle/
2 TYPE BIRTH
2 GIVN Kristopher
2 SURN Kringle
2 SOUR @S234@
3 PAGE Bar: 1
0 TRLR`,
			ExpectedNames: []gedcom.PersonalName{
				{
					Payload: "Santa /Clause/",
					Given:   "Santa",
					Surname: "Clause",
					SourceCitations: []*gedcom.SourceCitation{
						{
							Xref: "@S123@",
							Page: "Foo: 1",
						},
					},
				},
				{
					Payload: "Kristopher /Kringle/",
					Given:   "Kristopher",
					Surname: "Kringle",
					SourceCitations: []*gedcom.SourceCitation{
						{
							Xref: "@S234@",
							Page: "Bar: 1",
						},
					},
				},
			},
		})
	})
}

func cmpPersonalName(t *testing.T, errMsgPrefix string, got, exp gedcom.PersonalName) {
	if got.Payload != exp.Payload {
		t.Errorf("%swrong Payload, got %q, exp %q", errMsgPrefix, got.Payload, exp.Payload)
	}
	if got.NamePrefix != exp.NamePrefix {
		t.Errorf("%swrong NamePrefix, got %q, exp %q", errMsgPrefix, got.NamePrefix, exp.NamePrefix)
	}
	if got.Given != exp.Given {
		t.Errorf("%swrong Given, got %q, exp %q", errMsgPrefix, got.Given, exp.Given)
	}
	if got.Nickname != exp.Nickname {
		t.Errorf("%swrong Nickname, got %q, exp %q", errMsgPrefix, got.Nickname, exp.Nickname)
	}
	if got.SurnamePrefix != exp.SurnamePrefix {
		t.Errorf("%swrong SurnamePrefix, got %q, exp %q", errMsgPrefix, got.SurnamePrefix, exp.SurnamePrefix)
	}
	if got.Surname != exp.Surname {
		t.Errorf("%swrong Surname, got %q, exp %q", errMsgPrefix, got.Surname, exp.Surname)
	}
	if got.NameSuffix != exp.NameSuffix {
		t.Errorf("%swrong NameSuffix, got %q, exp %q", errMsgPrefix, got.NameSuffix, exp.NameSuffix)
	}

	if len(got.SourceCitations) != len(exp.SourceCitations) {
		t.Errorf("%s; wrong number of SourceCitations; got %d, exp %d", errMsgPrefix, len(got.SourceCitations), len(exp.SourceCitations))
	} else {
		for i, got := range got.SourceCitations {
			errMsgPrefix := fmt.Sprintf("%s.SourceCitations[%d], ", errMsgPrefix, i)
			cmpSourceCitation(t, errMsgPrefix, got, exp.SourceCitations[i])
		}
	}
}
