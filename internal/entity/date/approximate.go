package date

import (
	"fmt"
	"strings"
)

// approximateTokens may help qualify dateApprox values. In the GEDCOM7 spec,
// the ABNF grammar for a dateApprox is:
//
//	(%s"ABT" / %s"CAL" / %s"EST") D date
//
// Additionally, a token "ABOUT" is expanded to mean the same as "ABT". This is
// actually not allowed by the GEDCOM7 spec, but because vendors let you get
// away with it, it is accepted by this package for compatibility reasons.
var approximateTokens = []string{
	"ABT", // "ABT x" means exact date unknown, but near x.
	"CAL", // "CAL x" means x is calculated from other data.
	"EST", // "EST x" means exact date unknown, but near x; and x is calculated from other data.

	"ABOUT", // not officially GEDCOM7, but is accepted to mean the same thing as "ABT".
}

func parseDateWithApproximationToken(layout, in string) (out *Date, err error) {
	_, touchedIn, originallyApproximation, err := originallyApproximate(in)
	if err != nil {
		return nil, err
	}

	if out, err = parse(layout, touchedIn); err != nil {
		return nil, err
	}
	out.Payload = in

	if originallyApproximation {
		out.Approximate = true
	}
	out.setDisplay()
	return out, nil
}

// originallyApproximate determines if the input in specifically had a word,
// such as "ABT", "CAL", or "EST", to indicate that the date is an
// approximation of a date.
func originallyApproximate(in string) (token, date string, approx bool, err error) {
	uin := strings.ToUpper(in)

	for _, prefix := range approximateTokens {
		// GEDCOM7 specifies upper-case strings. But in practice, there is a lot
		// of data cased like "This", or "this". Accept any casing here so that
		// the application can work with that data.
		if strings.HasPrefix(uin, prefix) {
			token, date, approx = strings.Cut(in, " ")
			if len(date) < 1 {
				err = fmt.Errorf("malformatted approximate date %q", in)
				return
			}
			if approx {
				return
			}
		}
	}

	date = in
	return
}
