package date

import (
	"errors"
	"strings"
)

// Range expresses a time period with known or uncertain bounds. This type
// represents two possible symbols from the GEDCOM7 spec:
//
// The DatePeriod
//
//	[ %s"TO" D date ] /  %s"FROM" D date [ D %s"TO" D date ]
//
// And the dateRange
//
//	%s"BET" D date D %s"AND" D date / %s"AFT" D date / %s"BEF" D date
//
// Like interpretation of Date, interpretation of Range is looser with regards
// to approximation words (such as "FROM", "TO", "BET", "AND", "BEF", "AFT"), in
// that it is case-insensitive and will allow for expanded versions of said
// approximation words.
type Range struct {
	Lo, Hi  *Date
	Payload string
}

var errNotRange = errors.New("input does not appear to be a range at all")

const (
	rangeTokenTo   = "TO"
	rangeTokenFrom = "FROM"
	rangeTokenBet  = "BET"
	rangeTokenAnd  = "AND"
	rangeTokenAft  = "AFT"
	rangeTokenBef  = "BEF"
)

// parseRange attempts to interpret in as a Range. If it appears that the input
// is not a range in the first place, then err is errNotRange.
func parseRange(in string) (out *Range, err error) {
	out = &Range{}
	uin := strings.ToUpper(in)

	// First, attempt to interpret like a GEDCOM7 "DatePeriod"
	toInd := strings.Index(uin, rangeTokenTo)
	if toInd == 0 {
		out.Hi, err = parseDateWithoutApproximationToken(in[len(rangeTokenTo)+1:])
		if err == nil {
			out.Payload = in
		}
		return
	}

	startsWithFrom := strings.HasPrefix(uin, rangeTokenFrom)
	if startsWithFrom && toInd < 0 {
		out.Lo, err = parseDateWithoutApproximationToken(in[len(rangeTokenFrom)+1:])
		if err == nil {
			out.Payload = in
		}
		return
	} else if startsWithFrom && toInd > 0 {
		if out.Lo, err = parseDateWithoutApproximationToken(in[len(rangeTokenFrom)+1 : toInd-1]); err != nil {
			return
		}

		out.Hi, err = parseDateWithoutApproximationToken(in[toInd+len(rangeTokenTo)+1:])
		if err == nil {
			out.Payload = in
		}
		return
	}

	// Attempt to interpret like a GEDCOM7 "dateRange"
	andInd := strings.Index(uin, rangeTokenAnd)
	for _, token := range []string{rangeTokenBet + "WEEN", rangeTokenBet + ".", rangeTokenBet} {
		if strings.HasPrefix(uin, token) && andInd > 0 {
			if out.Lo, err = parseDateWithoutApproximationToken(in[len(token)+1 : andInd-1]); err != nil {
				return
			}

			out.Hi, err = parseDateWithoutApproximationToken(in[andInd+4:])
			if err == nil {
				out.Payload = in
			}

			return
		}
	}

	for _, token := range []string{rangeTokenAft + "ER", rangeTokenAft + ".", rangeTokenAft} {
		if strings.HasPrefix(uin, token) {
			out.Lo, err = parseDateWithoutApproximationToken(in[len(token)+1:])
			if err == nil {
				out.Payload = in
			}
			return
		}
	}

	for _, token := range []string{rangeTokenBef + "ORE", rangeTokenBef + ".", rangeTokenBef} {
		if strings.HasPrefix(uin, token) {
			out.Hi, err = parseDateWithoutApproximationToken(in[len(token)+1:])
			if err == nil {
				out.Payload = in
			}
			return
		}
	}

	err = errNotRange
	return
}

func parseDateWithoutApproximationToken(in string) (out *Date, err error) {
	errs := make([]error, 0, len(dateLayouts))
	for _, layout := range dateLayouts {
		out, err = parse(layout, in)
		if err != nil {
			errs = append(errs, err)
		} else {
			out.setDisplay()
			return
		}
	}

	err = errors.Join(errs...)
	return
}
