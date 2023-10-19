package gedcom

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var dateLayouts = []string{
	"2 Jan 2006",
	"Jan 2006",
	"2 January 2006",
	"January 2006",
	"2006",
}

func newDate(in string) (out *time.Time, err error) {
	var val time.Time

	if strings.HasPrefix(strings.ToLower(in), "abt") {
		fields := strings.Fields(in)

		if len(fields) < 2 {
			return nil, fmt.Errorf("malformatted approximate date %q", in)
		}
		in = fields[1]
	}

	errs := make([]error, 0, len(dateLayouts))
	for _, layout := range dateLayouts {
		val, err = time.Parse(layout, in)
		if err == nil {
			out = &val
			return
		}
		errs = append(errs, err)
	}

	err = errors.Join(errs...)
	return
}
