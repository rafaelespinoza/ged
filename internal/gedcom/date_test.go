package gedcom

import (
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	tests := []struct {
		In     string
		Exp    time.Time
		ExpErr bool
	}{
		{
			In:  "ABT 1688",
			Exp: time.Date(1688, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			In:  "2 JAN 2006",
			Exp: time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			In:  "Jan 2006",
			Exp: time.Date(2006, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			In:  "2 January 2006",
			Exp: time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			In:  "2006",
			Exp: time.Date(2006, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			In:     "bad",
			ExpErr: true,
		},
		{
			In:     "abt",
			ExpErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.In, func(t *testing.T) {
			got, err := newDate(test.In)
			if err != nil && !test.ExpErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && test.ExpErr {
				t.Fatalf("expected an error but got %v", err)
			} else if err != nil && test.ExpErr {
				t.Log(err) // inspect error manually for now
				return
			}

			if got == nil {
				t.Fatal("output should be non-empty")
			}

			if !got.Equal(test.Exp) {
				t.Errorf("wrong value; got %q, expected %q", got.Format(time.RFC3339), test.Exp.Format(time.RFC3339))
			}
		})
	}
}
