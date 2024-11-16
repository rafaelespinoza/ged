package date_test

import (
	"testing"

	"github.com/rafaelespinoza/ged/internal/entity/date"
)

func TestCmpDates(t *testing.T) {
	tests := []struct {
		Name     string
		InputA   *date.Date
		InputB   *date.Date
		Expected int
	}{
		{
			Name:     "a.Year < b.Year",
			InputA:   mustParseDate(t, "1 Jan 2023"),
			InputB:   mustParseDate(t, "1 Jan 2024"),
			Expected: -1,
		},
		{
			Name:     "a.Year > b.Year",
			InputA:   mustParseDate(t, "1 Jan 2024"),
			InputB:   mustParseDate(t, "1 Jan 2023"),
			Expected: 1,
		},
		{
			Name:     "a.Month < b.Month",
			InputA:   mustParseDate(t, "1 Jan 2024"),
			InputB:   mustParseDate(t, "1 Feb 2024"),
			Expected: -1,
		},
		{
			Name:     "a.Month > b.Month",
			InputA:   mustParseDate(t, "1 Feb 2024"),
			InputB:   mustParseDate(t, "1 Jan 2024"),
			Expected: 1,
		},
		{
			Name:     "a.Day < b.Day",
			InputA:   mustParseDate(t, "1 Feb 2024"),
			InputB:   mustParseDate(t, "2 Feb 2024"),
			Expected: -1,
		},
		{
			Name:     "a.Day > b.Day",
			InputA:   mustParseDate(t, "2 Feb 2024"),
			InputB:   mustParseDate(t, "1 Feb 2024"),
			Expected: 1,
		},
		{
			Name:     "equal",
			InputA:   mustParseDate(t, "1 Feb 2024"),
			InputB:   mustParseDate(t, "1 Feb 2024"),
			Expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := date.CmpDates(test.InputA, test.InputB)

			if got != test.Expected {
				t.Errorf("got %d, expected %d", got, test.Expected)
			}
		})
	}
}

func TestCmpDateRanges(t *testing.T) {
	emptyDateRange := &date.Range{}

	tests := []struct {
		Name     string
		InputA   *date.Range
		InputB   *date.Range
		Expected int
	}{
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo non-empty, b.Hi non-empty; a.Lo < b.Lo",
			InputA:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "From 2 Jan 2024 To 3 Jan 2024"),
			Expected: -1,
		},
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo non-empty, b.Hi non-empty; a.Lo == b.Lo",
			InputA:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			Expected: 0,
		},
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo non-empty, b.Hi non-empty; a.Lo > b.Lo",
			InputA:   mustParseRange(t, "From 2 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			Expected: 1,
		},

		// permutations on InputA, InputB is always non-empty
		{
			Name:     "a.Lo non-empty, a.Hi empty; b.Lo non-empty, b.Hi non-empty; a.Lo < b.Lo",
			InputA:   mustParseRange(t, "From 1 Jan 2024"),
			InputB:   mustParseRange(t, "From 2 Jan 2024 To 3 Jan 2024"),
			Expected: -1,
		},
		{
			Name:     "a.Lo non-empty, a.Hi empty; b.Lo non-empty, b.Hi non-empty; a.Lo == b.Lo",
			InputA:   mustParseRange(t, "From 1 Jan 2024"),
			InputB:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			Expected: 0,
		},
		{
			Name:     "a.Lo non-empty, a.Hi empty; b.Lo non-empty, b.Hi non-empty; a.Lo > b.Lo",
			InputA:   mustParseRange(t, "From 2 Jan 2024"),
			InputB:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			Expected: 1,
		},
		{
			Name:     "a.Lo empty, a.Hi non-empty; b.Lo non-empty, b.Hi non-empty; a.Hi < b.Lo",
			InputA:   mustParseRange(t, "To 1 Jan 2024"),
			InputB:   mustParseRange(t, "From 2 Jan 2024 To 3 Jan 2024"),
			Expected: -1,
		},
		{
			Name:     "a.Lo empty, a.Hi non-empty; b.Lo non-empty, b.Hi non-empty; a.Hi == b.Lo",
			InputA:   mustParseRange(t, "To 2 Jan 2024"),
			InputB:   mustParseRange(t, "From 2 Jan 2024 To 3 Jan 2024"),
			Expected: 0,
		},
		{
			Name:     "a.Lo empty, a.Hi non-empty; b.Lo non-empty, b.Hi non-empty; a.Lo > b.Lo",
			InputA:   mustParseRange(t, "To 2 Jan 2024"),
			InputB:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			Expected: 1,
		},
		{
			Name:     "a empty; b non-empty",
			InputA:   emptyDateRange,
			InputB:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			Expected: 1,
		},

		// permutations on InputB, InputA is always non-empty
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo empty, b.Hi non-empty; a.Lo < b.Hi",
			InputA:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "To 3 Jan 2024"),
			Expected: -1,
		},
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo empty, b.Hi non-empty; a.Lo == b.Hi",
			InputA:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "To 1 Jan 2024"),
			Expected: 0,
		},
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo empty, b.Hi non-empty; a.Lo > b.Hi",
			InputA:   mustParseRange(t, "From 2 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "From 1 Jan 2024"),
			Expected: 1,
		},
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo non-empty, b.Hi empty; a.Lo < b.Lo",
			InputA:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "From 2 Jan 2024"),
			Expected: -1,
		},
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo non-empty, b.Hi empty; a.Lo == b.Lo",
			InputA:   mustParseRange(t, "From 2 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "From 2 Jan 2024"),
			Expected: 0,
		},
		{
			Name:     "a.Lo non-empty, a.Hi non-empty; b.Lo non-empty, b.Hi empty; a.Lo > b.Lo",
			InputA:   mustParseRange(t, "From 2 Jan 2024 To 3 Jan 2024"),
			InputB:   mustParseRange(t, "From 1 Jan 2024"),
			Expected: 1,
		},
		{
			Name:     "a non-empty; b empty",
			InputA:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			InputB:   emptyDateRange,
			Expected: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := date.CmpDateRanges(test.InputA, test.InputB)

			if got != test.Expected {
				t.Errorf("got %d, expected %d", got, test.Expected)
			}
		})
	}
}

func TestCmpDateToDateRange(t *testing.T) {
	tests := []struct {
		Name     string
		InputA   *date.Date
		InputB   *date.Range
		Expected int
	}{
		{
			Name:     "a non-nil; b nil",
			InputA:   mustParseDate(t, "1 Jan 2024"),
			InputB:   nil,
			Expected: -1,
		},
		{
			Name:     "a empty; b nil",
			InputA:   nil,
			InputB:   nil,
			Expected: 0,
		},
		{
			Name:     "a nil; b non-nil",
			InputA:   nil,
			InputB:   mustParseRange(t, "From 1 Jan 2024 To 3 Jan 2024"),
			Expected: 1,
		},
		{
			Name:     "a non-nil; b.Lo non-nil, b.Hi nil",
			InputA:   mustParseDate(t, "1 Jan 2024"),
			InputB:   mustParseRange(t, "From 1 Jan 2024"),
			Expected: 0,
		},
		{
			Name:     "a non-nil; b.Lo nil, b.Hi non-nil",
			InputA:   mustParseDate(t, "1 Jan 2024"),
			InputB:   mustParseRange(t, "To 1 Jan 2024"),
			Expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := date.CmpDateToDateRange(test.InputA, test.InputB)

			if got != test.Expected {
				t.Errorf("got %d, expected %d", got, test.Expected)
			}
		})
	}
}
