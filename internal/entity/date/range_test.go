package date_test

import (
	"testing"
	"time"

	"github.com/rafaelespinoza/ged/internal/entity/date"
)

func TestParseRange(t *testing.T) {
	type Testcase struct {
		Name        string
		Input       string
		Expected    *date.Range
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
			if rng != nil {
				t.Fatal("Range output expected to be empty when there's an error")
			}
			return
		}
		if dat != nil {
			t.Fatal("Date output expected to be empty when the input is otherwise valid for a Range")
		}

		exp := test.Expected

		if rng.Lo != nil && exp.Lo == nil {
			t.Fatalf("expected empty value for Lo, but got %v", rng.Lo)
		} else if rng.Lo == nil && exp.Lo != nil {
			t.Fatalf("expected non-empty value for Lo, but got %v", rng.Lo)
		} else if rng.Lo != nil && exp.Lo != nil {
			testDate(t, rng.Lo, exp.Lo)
		}

		if rng.Hi != nil && exp.Hi == nil {
			t.Fatalf("expected empty value for Hi, but got %v", rng.Hi)
		} else if rng.Hi == nil && exp.Hi != nil {
			t.Fatalf("expected non-empty value for Hi, but got %v", rng.Hi)
		} else if rng.Hi != nil && exp.Hi != nil {
			testDate(t, rng.Hi, exp.Hi)
		}

		if rng.Payload != exp.Payload {
			t.Errorf("wrong Payload; got %q, expected %q", rng.Payload, exp.Payload)
		}
	}

	t.Run("TO", func(t *testing.T) {
		tests := []Testcase{
			{
				Name:  "all parts present",
				Input: "2 Jan 2006",
				Expected: &date.Range{
					Hi: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
				},
			},
			{
				Name:  "all parts present, month fully spelled out",
				Input: "2 January 2006",
				Expected: &date.Range{
					Hi: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
				},
			},
			{
				Name:  "only month year",
				Input: "Jan 2006",
				Expected: &date.Range{
					Hi: &date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
				},
			},
			{
				Name:  "only year",
				Input: "2006",
				Expected: &date.Range{
					Hi: &date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				for _, token := range []string{"TO", "To"} {
					tokenizedInput := token + " " + test.Input

					t.Run(test.Name, func(t *testing.T) {
						originalInput := test.Input
						defer func() { test.Input = originalInput }()
						test.Input = tokenizedInput
						runTest(t, test)
					})
				}
			})
		}
	})

	t.Run("FROM ... TO", func(t *testing.T) {
		tests := []Testcase{
			{
				Name:  "all parts present, tokens all caps",
				Input: "FROM 2 Jan 2006 TO 17 Jan 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
					Hi: &date.Date{Year: 2038, Month: time.January, Day: 17, Display: "2038-01-17"},
				},
			},
			{
				Name:  "all parts present, tokens mixed case",
				Input: "From 2 Jan 2006 To 17 Jan 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
					Hi: &date.Date{Year: 2038, Month: time.January, Day: 17, Display: "2038-01-17"},
				},
			},
			{
				Name:  "all parts present, month fully spelled out, tokens all caps",
				Input: "FROM 2 January 2006 TO 17 January 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
					Hi: &date.Date{Year: 2038, Month: time.January, Day: 17, Display: "2038-01-17"},
				},
			},
			{
				Name:  "all parts present, month fully spelled out, tokens mixed case",
				Input: "From 2 January 2006 To 17 January 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
					Hi: &date.Date{Year: 2038, Month: time.January, Day: 17, Display: "2038-01-17"},
				},
			},
			{
				Name:  "only month year, tokens all caps",
				Input: "FROM Jan 2006 TO Jan 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
					Hi: &date.Date{Year: 2038, Month: time.January, Approximate: true, Display: "~ 2038-01"},
				},
			},
			{
				Name:  "only month year, tokens mixed case",
				Input: "From Jan 2006 To Jan 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
					Hi: &date.Date{Year: 2038, Month: time.January, Approximate: true, Display: "~ 2038-01"},
				},
			},
			{
				Name:  "only year, tokens all caps",
				Input: "FROM 2006 TO 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
					Hi: &date.Date{Year: 2038, Approximate: true, Display: "~ 2038"},
				},
			},
			{
				Name:  "only year, tokens mixed cased",
				Input: "From 2006 To 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
					Hi: &date.Date{Year: 2038, Approximate: true, Display: "~ 2038"},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				runTest(t, test)
			})
		}
	})

	t.Run("FROM", func(t *testing.T) {
		tests := []Testcase{
			{
				Name:  "all parts present",
				Input: "2 Jan 2006",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
				},
			},
			{
				Name:  "all parts present, month fully spelled out",
				Input: "2 January 2006",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
				},
			},
			{
				Name:  "only month year",
				Input: "Jan 2006",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
				},
			},
			{
				Name:  "only year",
				Input: "2006",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				for _, token := range []string{"FROM", "From"} {
					tokenizedInput := token + " " + test.Input

					t.Run(token, func(t *testing.T) {
						originalInput := test.Input
						defer func() { test.Input = originalInput }()
						test.Input = tokenizedInput
						runTest(t, test)
					})
				}
			})
		}
	})

	t.Run("BET ... AND", func(t *testing.T) {
		tests := []Testcase{
			{
				Name:  "all parts present, AND token all caps",
				Input: "2 Jan 2006 AND 17 Jan 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
					Hi: &date.Date{Year: 2038, Month: time.January, Day: 17, Display: "2038-01-17"},
				},
			},
			{
				Name:  "all parts present, And token mixed case",
				Input: "2 Jan 2006 And 17 Jan 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
					Hi: &date.Date{Year: 2038, Month: time.January, Day: 17, Display: "2038-01-17"},
				},
			},
			{
				Name:  "all parts present, month fully spelled out, AND token all caps",
				Input: "2 January 2006 AND 17 January 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
					Hi: &date.Date{Year: 2038, Month: time.January, Day: 17, Display: "2038-01-17"},
				},
			},
			{
				Name:  "all parts present, month fully spelled out, And token mixed case",
				Input: "2 January 2006 And 17 January 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
					Hi: &date.Date{Year: 2038, Month: time.January, Day: 17, Display: "2038-01-17"},
				},
			},
			{
				Name:  "only month year, AND token all caps",
				Input: "Jan 2006 AND Jan 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
					Hi: &date.Date{Year: 2038, Month: time.January, Approximate: true, Display: "~ 2038-01"},
				},
			},
			{
				Name:  "only month year, tokens all caps, And token mixed case",
				Input: "Jan 2006 And Jan 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
					Hi: &date.Date{Year: 2038, Month: time.January, Approximate: true, Display: "~ 2038-01"},
				},
			},
			{
				Name:  "only year, AND token all caps",
				Input: "2006 AND 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
					Hi: &date.Date{Year: 2038, Approximate: true, Display: "~ 2038"},
				},
			},
			{
				Name:  "only year, tokens all caps, And token mixed case",
				Input: "2006 And 2038",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
					Hi: &date.Date{Year: 2038, Approximate: true, Display: "~ 2038"},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				for _, token := range []string{"BET", "Bet", "Bet.", "Between"} {
					tokenizedInput := token + " " + test.Input

					t.Run(token, func(t *testing.T) {
						originalInput := test.Input
						defer func() { test.Input = originalInput }()
						test.Input = tokenizedInput
						runTest(t, test)
					})
				}
			})
		}
	})

	t.Run("AFT", func(t *testing.T) {
		tests := []Testcase{
			{
				Name:  "all parts present",
				Input: "2 Jan 2006",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
				},
			},
			{
				Name:  "all parts present, month fully spelled out",
				Input: "2 January 2006",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
				},
			},
			{
				Name:  "only month year",
				Input: "Jan 2006",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
				},
			},
			{
				Name:  "only year",
				Input: "2006",
				Expected: &date.Range{
					Lo: &date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				for _, token := range []string{"AFT", "Aft", "Aft.", "After"} {
					tokenizedInput := token + " " + test.Input

					t.Run(token, func(t *testing.T) {
						originalInput := test.Input
						defer func() { test.Input = originalInput }()
						test.Input = tokenizedInput
						runTest(t, test)
					})
				}
			})
		}
	})

	t.Run("BEF", func(t *testing.T) {
		tests := []Testcase{
			{
				Name:  "all parts present",
				Input: "2 Jan 2006",
				Expected: &date.Range{
					Hi: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
				},
			},
			{
				Name:  "all parts present, month fully spelled out",
				Input: "2 January 2006",
				Expected: &date.Range{
					Hi: &date.Date{Year: 2006, Month: time.January, Day: 2, Display: "2006-01-02"},
				},
			},
			{
				Name:  "only month year",
				Input: "Jan 2006",
				Expected: &date.Range{
					Hi: &date.Date{Year: 2006, Month: time.January, Approximate: true, Display: "~ 2006-01"},
				},
			},
			{
				Name:  "only year",
				Input: "2006",
				Expected: &date.Range{
					Hi: &date.Date{Year: 2006, Approximate: true, Display: "~ 2006"},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				for _, token := range []string{"BEF", "Bef", "Bef.", "Before"} {
					tokenizedInput := token + " " + test.Input

					t.Run(token, func(t *testing.T) {
						originalInput := test.Input
						defer func() { test.Input = originalInput }()
						test.Input = tokenizedInput
						runTest(t, test)
					})
				}
			})
		}
	})
}
