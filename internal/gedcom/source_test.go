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

			if len(actual.Components) != len(expected.Components) {
				t.Errorf("wrong number of Components; got %d, exp %d", len(actual.Components), len(expected.Components))
			} else {
				for i, got := range actual.Components {
					exp := expected.Components[i]
					if got != exp {
						t.Errorf("item[%d]; got %q, exp %q", i, got, exp)
					}
				}
			}

			if len(actual.Tuples) != len(expected.Tuples) {
				t.Errorf("wrong number of Tuples; got %d, exp %d", len(actual.Tuples), len(expected.Tuples))
			} else {
				for key, exp := range expected.Tuples {
					got, ok := actual.Tuples[key]
					if !ok {
						t.Errorf("expected a key %q, but not found", key)
					} else if got != exp {
						t.Errorf("item[%q]; got %q, exp %q", key, got, exp)
					}
				}

				for key, got := range actual.Tuples {
					exp, ok := expected.Tuples[key]
					if !ok {
						t.Errorf("unexpected key %q", key)
					} else if got != exp {
						t.Errorf("item[%q]; got %q, exp %q", key, got, exp)
					}
				}
			}
		})
	}
}
