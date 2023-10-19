package gedcom

import (
	"fmt"
	"time"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"
)

// Event is a general-purpose struct that could be used to record events in an
// individual's life, or a family history. Its URI is g7:INDI-EVEN. This type
// could probably serve as the basis for more specific events, such as BIRT (to
// record the birth of an individual), or MARR (to record the start of a
// marriage in a family). See the GEDCOM7 spec for info on Individual Events and
// Family Events.
type Event struct {
	Date  *time.Time
	Place string
}

func newEvent(line *gedcom7.Line, subnodes []*gedcom.Node) (out *Event, err error) {
	var (
		subline *gedcom7.Line
		date    *time.Time
		place   string
	)

	for _, subnode := range subnodes {
		if subline, err = parseLine(subnode); err != nil {
			return
		}

		switch subline.Tag {
		case "DATE":
			if date != nil {
				err = fmt.Errorf("error parsing event, multiple DATE lines, conflicting line: %q", subline.Text)
				return
			}

			date, err = newDate(subline.Payload)
			if err != nil {
				return
			}
		case "PLAC":
			if place != "" {
				err = fmt.Errorf("error parsing event, multiple PLAC lines, conflicting line: %q", subline.Text)
				return
			}

			place = subline.Payload
		}
	}

	out = &Event{Date: date, Place: place}
	return
}
