package gedcom_test

import (
	"testing"

	"github.com/rafaelespinoza/ged/internal/gedcom"
)

func TestSourceCitationPage(t *testing.T) {
	tests := []struct {
		Name           string
		Citation       gedcom.SourceCitation
		FieldDelimiter string
		Expected       gedcom.SourcePage
		ExpectError    bool
	}{
		{
			Name: "no Tuples",
			Citation: gedcom.SourceCitation{
				Page: "The General Civil Archive, Springfield; Springfield, USA",
			},
			FieldDelimiter: ";",
			Expected: gedcom.SourcePage{
				Payload:    "The General Civil Archive, Springfield; Springfield, USA",
				Components: []string{"The General Civil Archive, Springfield", "Springfield, USA"},
			},
		},
		{
			Name: "with Tuples",
			Citation: gedcom.SourceCitation{
				Page: "Year: 1880; Census Place: Hill Valley, Hill County, California; Page: 2B; Enumeration District: 1234; FHL microfilm: 1111111",
			},
			FieldDelimiter: ";",
			Expected: gedcom.SourcePage{
				Payload: "Year: 1880; Census Place: Hill Valley, Hill County, California; Page: 2B; Enumeration District: 1234; FHL microfilm: 1111111",
				Tuples: map[string]string{
					"Year":                 "1880",
					"Census Place":         "Hill Valley, Hill County, California",
					"Page":                 "2B",
					"Enumeration District": "1234",
					"FHL microfilm":        "1111111",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			actual, err := test.Citation.ParsePage(test.FieldDelimiter)

			if test.ExpectError && err == nil {
				t.Fatal("expected an error but got nil")
			} else if !test.ExpectError && err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			expected := test.Expected
			if actual.Payload != expected.Payload {
				t.Errorf("wrong Payload; got %q, exp %q", actual.Payload, expected.Payload)
			}

			cmpStringSlices(t, "", actual.Components, expected.Components)
			cmpStringMaps(t, "", actual.Tuples, expected.Tuples)
		})
	}
}

func cmpSourceCitation(t *testing.T, errMsgPrefix string, got, exp *gedcom.SourceCitation) {
	if got.Xref != exp.Xref {
		t.Errorf("%swrong Xref; got %q, exp %q", errMsgPrefix, got.Xref, exp.Xref)
	}
	if got.Page != exp.Page {
		t.Errorf("%swrong Page; got %q, exp %q", errMsgPrefix, got.Page, exp.Page)
	}

	cmpStringMaps(t, errMsgPrefix, got.Data, exp.Data)
}
