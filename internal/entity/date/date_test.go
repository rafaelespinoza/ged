package date_test

import (
	"testing"
	"time"

	"github.com/rafaelespinoza/ged/internal/entity/date"
)

func TestParseDate(t *testing.T) {
	type Testcase struct {
		Name        string
		Input       string
		Expected    date.Date
		ExpectError bool
	}

	runTest := func(t *testing.T, test Testcase) {
		test.Expected.Payload = test.Input

		dat, rng, err := date.Parse(test.Input)
		if err != nil && !test.ExpectError {
			t.Fatal(err)
		} else if err == nil && test.ExpectError {
			t.Fatal("expected error but got nil")
		} else if err != nil && test.ExpectError {
			if dat != nil {
				t.Fatal("Date output expected to be empty when there's an error")
			}
			return
		}
		if rng != nil {
			t.Logf("%#v", rng)
			t.Fatal("Range output expected to be empty when the input is otherwise valid for a Date")
		}

		testDate(t, dat, &test.Expected)
	}

	t.Run("no explicit approximation", func(t *testing.T) {
		// [calendar D] [[day D] month D] year [D epoch]
		tests := []Testcase{
			{
				Name:     "all parts present",
				Input:    "2 Jan 2006",
				Expected: date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
			},
			{
				Name:     "all parts present, month fully spelled out",
				Input:    "2 January 2006",
				Expected: date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
			},
			{
				Name:     "only month year",
				Input:    "Jan 2006",
				Expected: date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
			},
			{
				Name:     "only year",
				Input:    "2006",
				Expected: date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
		}
	})

	t.Run("explicit approximation", func(t *testing.T) {
		tests := []Testcase{
			{
				Name:     "all parts present",
				Input:    "2 Jan 2006",
				Expected: date.Date{Year: 2006, Month: time.January, Day: 2, Approximate: true, Display: "~ 2006-01-02"},
			},
			{
				Name:     "all parts present, month fully spelled out",
				Input:    "2 January 2006",
				Expected: date.Date{Year: 2006, Month: time.January, Day: 2, Approximate: true, Display: "~ 2006-01-02"},
			},
			{
				Name:     "only month year",
				Input:    "Jan 2006",
				Expected: date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
			},
			{
				Name:     "only year",
				Input:    "2006",
				Expected: date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
			},
		}

		for _, test := range tests {
			tokens := []string{
				"ABT", "ABT.", "Abt", "Abt.", "abt", "abt.", "About",
				"CAL",
				"EST", "EST.", "Est", "Est.", "est", "est.",
			}

			for _, approxToken := range tokens {
				approximatedInput := approxToken + " " + test.Input

				t.Run(approximatedInput, func(t *testing.T) {
					originalInput := test.Input
					defer func() { test.Input = originalInput }()
					test.Input = approximatedInput

					runTest(t, test)
				})
			}
		}
	})
}

func testDate(t *testing.T, got, exp *date.Date) {
	t.Helper()

	if got.Year != exp.Year {
		t.Errorf("wrong Year; got %d, expected %d", got.Year, exp.Year)
	}
	if got.Month != exp.Month {
		t.Errorf("wrong Month; got %d, expected %d", got.Month, exp.Month)
	}
	if got.Day != exp.Day {
		t.Errorf("wrong Day; got %d, expected %d", got.Day, exp.Day)
	}
	if got.Approximate != exp.Approximate {
		t.Errorf("wrong Approximate; got %t, expected %t", got.Approximate, exp.Approximate)
	}
	if got.Payload != exp.Payload {
		t.Errorf("wrong Payload; got %q, expected %q", got.Payload, exp.Payload)
	}
	if got.Display != exp.Display {
		t.Errorf("wrong Display; got %q, expected %q", got.Display, exp.Display)
	}
}

func mustParseDate(t *testing.T, in string) (out *date.Date) {
	t.Helper()

	out, rng, err := date.Parse(in)
	if err != nil {
		t.Fatalf("bad input %q: %v", in, err)
	} else if rng != nil {
		t.Fatalf("input %q appears to be a *Range, not a *Date", in)
	}

	return
}

func mustParseRange(t *testing.T, in string) (out *date.Range) {
	t.Helper()

	dat, out, err := date.Parse(in)
	if err != nil {
		t.Fatalf("bad input %q: %v", in, err)
	} else if dat != nil {
		t.Fatalf("input %q appears to be a *Date, not a *Range", in)
	}

	return
}
