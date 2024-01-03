package entity

import (
	"errors"

	"github.com/rafaelespinoza/ged/internal/entity/date"
)

// Date is meant to approximate a union type in that the value of interest could
// be either the Date or the Range, but not both. It probably cannot be a
// generic type, such as: `interface { date.Date | date.Range }` and also be
// part of some essential container types, such as Person or Union, because it
// would impose unacceptable constraints. For instance, A Person with a
// Birthdate of type *date.Date would only be able able to have Children and
// Parents whose Birthdate field is also of type *date.Date; such relatives with
// Birthdate field of type *date.Range would not be allowed.
type Date struct {
	*date.Date
	*date.Range
}

func NewDate(d *date.Date, r *date.Range) (out *Date, err error) {
	if d != nil && r != nil {
		err = errors.New("invalid NewDate, Date and Range cannot both be non-empty")
	} else {
		out = &Date{Date: d, Range: r}
	}
	return
}

func (d Date) String() string {
	var out string

	if d.Date != nil {
		out = d.Date.Display
	} else if d.Range != nil {
		if d.Range.Lo != nil && d.Range.Hi != nil {
			out = d.Range.Lo.Display + " ... " + d.Range.Hi.Display
		} else if d.Range.Lo != nil {
			out = ">= " + d.Range.Lo.Display
		} else if d.Range.Hi != nil {
			out = "<= " + d.Range.Hi.Display
		}
	} else {
		out = "?"
	}

	return out
}
