// Package date provides a simple programming interface for GEDCOM-formatted dates.
// Read more GEDCOM at https://gedcom.io.
//
// This implementation is designed to shape data as described in the GEDCOM v7
// specification, but is a bit looser than that specification so that it may
// work with older GEDCOM data from different vendors. In particular, for the
// interpretation of approximation tokens ABT, CAL, EST for individual dates;
// and tokens FROM, TO, BET, AND, BEF, AFT for date ranges; are case-insensitive
// in this package, whereas a strict GEDCOM v7 interpretation would only
// recognize upper-cased tokens. Another way that approximation tokens here are
// looser than the actual GEDCOM v7 spec is that alternative words for certain
// are accepted. Those are noted elsewhere in this package's documentation.
package date

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Parse interprets in as a GEDCOM7-formatted date string and will return one of
// the following:
//   - *Date, nil, nil
//   - nil, *Range, nil
//   - nil, nil, error
func Parse(in string) (*Date, *Range, error) {
	if len(in) > 100 {
		return nil, nil, fmt.Errorf("max length is 100, but input is length %d", len(in))
	}

	rng, err := parseRange(in)
	if errors.Is(err, errNotRange) {
		// attempt another interpretation
	} else if err != nil {
		return nil, nil, err
	} else {
		return nil, rng, nil
	}

	dat, err := parseDate(in)
	return dat, nil, err
}

// Date is a general-purpose date with the maximum resolution of 1 day. It's
// meant to represent either a fully-known date, or an approximation of a date.
// This status is indicated by the field, Approximate. From the GEDCOM7
// specification, this type represents date, dateApprox, and DateExact. Only the
// Gregorian calendar is supported.
//
// Some ABNF snippets from the GEDCOM7 spec. But first, a brief primer on ABNF
// notation.
//
//	D is character 0x20 (one space).
//	[ ] mark optional components.
//	( ) mark a grouping of components.
//	"" double quotes mean string literals.
//	%s marks a case-sensitive string literal.
//	/ delimits alternatives.
//
// Back to official GEDCOM dates in ABNF notation:
//
//	date		= [calendar D] [[day D] month D] year [D epoch]
//	dateApprox	= (%s"ABT" / %s"CAL" / %s"EST") D date
//	DateExact	= day D month D year
//
// This package only considers the Gregorian calendar, and ignores the epoch
// (such as BCE). Also, it is looser with regards to approximation words (such
// as "ABT", "CAL", "EST"), in that it is case-insensitive and will allow for
// expanded versions of said approximation words. One example is that any of
// "Abt", "Abt.", "About", "ABT" are interpreted as the same thing. The
// actual interpretation in this package is more like the ABNF notation was:
//
//	date 		= [[day D] month] year
//	dateApprox	= (about / calculated / estimated) date
//	DateExact	= day D month D year
//	about		= (ABT / Abt. / about)
//	calculated	= (CAL / Cal. / calculated)
//	estimated	= (EST / Est. / estimated)
type Date struct {
	// Year is from the Gregorian calendar.
	Year int
	// Month is from the Gregorian calendar. It is 0 if the GEDCOM input did
	// not specify a month in the first place.
	Month time.Month
	// Day is the day of the month, if specified by the GEDCOM data. It would be
	// 0 if the GEDCOM input did not specify a day of the month.
	Day int
	// Approximate indicates if the date should be considered inexact or not.
	// If the original GEDCOM data explicitly said that the date was ABT
	// (about), CAL (calculated), or EST (estimated), then this will be true.
	// Another way this can be true is if the original input omitted the day of
	// the month, or omitted both the day of the month and the month.
	Approximate bool
	// Payload is the original input data from the GEDCOM node.
	Payload string
	// Display is an opinionated presentation of the date fields in a layout
	// that strives to be like YYYY-MM-DD using the available non-zero data and
	// indicating approximation with the prefix ~. It is not part of any GEDCOM
	// spec. Some example values:
	//
	//	2006-01-02
	//	~ 2006-01-02
	//	~ 2006-01
	//	~ 2006
	Display string
}

// setDisplay needs to be called after the Date has had all opportunities to
// check whether or not it is an Approximate date or not.
func (d *Date) setDisplay() {
	var b strings.Builder
	if d.Approximate {
		_, _ = b.WriteString("~ ")
	}

	parts := make([]string, 0, 3)
	parts = append(parts, strconv.Itoa(d.Year))

	var monthPart string
	if d.Month < time.January || d.Month > time.December {
		// no op
	} else if d.Month < time.October {
		monthPart = "0" + strconv.Itoa(int(d.Month))
	} else {
		monthPart = strconv.Itoa(int(d.Month))
	}

	if monthPart != "" {
		parts = append(parts, monthPart)

		if d.Day > 0 && d.Day < 10 {
			parts = append(parts, "0"+strconv.Itoa(int(d.Day)))
		} else if d.Day >= 10 {
			parts = append(parts, strconv.Itoa(int(d.Day)))
		}
	}
	_, _ = b.WriteString(strings.Join(parts, "-"))

	d.Display = b.String()
}

const (
	dateLayoutDMYAbbrevMonth = "2 Jan 2006"
	dateLayoutDMYFullMonth   = "2 January 2006"
	dateLayoutMYAbbrevMonth  = "Jan 2006"
	dateLayoutMYFullMonth    = "January 2006"
	dateLayoutY              = "2006"
)

var dateLayouts = []string{
	dateLayoutDMYAbbrevMonth,
	dateLayoutDMYFullMonth,
	dateLayoutMYAbbrevMonth,
	dateLayoutMYFullMonth,
	dateLayoutY,
}

func parseDate(in string) (out *Date, err error) {
	errs := make([]error, 0, len(dateLayouts))
	for _, layout := range dateLayouts {
		out, err = parseDateWithApproximationToken(layout, in)
		if err != nil {
			errs = append(errs, err)
		} else {
			return
		}
	}

	err = errors.Join(errs...)
	return
}

func parse(layout, in string) (*Date, error) {
	val, err := time.Parse(layout, in)
	if err != nil {
		return nil, err
	}

	y, m, d := val.Date()
	switch layout {
	case dateLayoutMYAbbrevMonth, dateLayoutMYFullMonth:
		d = 0
	case dateLayoutY:
		m, d = 0, 0
	default:
		break // no op
	}
	out := &Date{
		Year:  y,
		Month: m,
		Day:   d,
	}
	if out.Month == 0 || out.Day == 0 {
		out.Approximate = true
	}
	return out, nil
}
