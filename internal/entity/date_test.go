package entity_test

import (
	"testing"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/entity/date"
)

func TestDate(t *testing.T) {
	tests := []struct {
		Name      string
		InDate    *date.Date
		InRange   *date.Range
		ExpError  bool
		ExpString string
	}{
		{
			Name:      "both inputs are empty",
			ExpString: "?",
		},
		{
			Name:     "both inputs are non-empty",
			InDate:   &date.Date{},
			InRange:  &date.Range{},
			ExpError: true,
		},
		{
			Name:      "Date",
			InDate:    &date.Date{Display: "2006-01-02"},
			ExpString: "2006-01-02",
		},
		{
			Name: "Range: lo + hi",
			InRange: &date.Range{
				Lo: &date.Date{Display: "2006-01-02"},
				Hi: &date.Date{Display: "2038-01-19"},
			},
			ExpString: "2006-01-02 ... 2038-01-19",
		},
		{
			Name: "Range: lo",
			InRange: &date.Range{
				Lo: &date.Date{Display: "2006-01-02"},
			},
			ExpString: ">= 2006-01-02",
		},
		{
			Name: "Range: Hi",
			InRange: &date.Range{
				Hi: &date.Date{Display: "2038-01-19"},
			},
			ExpString: "<= 2038-01-19",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			dat, err := entity.NewDate(test.InDate, test.InRange)

			if err != nil && !test.ExpError {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && test.ExpError {
				t.Fatalf("expected error, but got %v", err)
			} else if err != nil && test.ExpError {
				return // ok
			}

			if dat == nil {
				t.Fatal("expected output to be non-empty")
			}

			got := dat.String()
			if got != test.ExpString {
				t.Errorf("wrong String; got %q, expected %q", got, test.ExpString)
			}
		})
	}
}
