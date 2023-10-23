package gedcom

import (
	"context"
	"fmt"
	"time"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"
	"github.com/rafaelespinoza/ged/internal/log"
)

// Event is a general-purpose struct that could be used to record events in an
// individual's life, or a family history. Its URI is g7:INDI-EVEN. This type
// could probably serve as the basis for more specific events, such as BIRT (to
// record the birth of an individual), or MARR (to record the start of a
// marriage in a family). See the GEDCOM7 spec for info on Individual Events and
// Family Events.
type Event struct {
	Date            *time.Time
	Place           string
	SourceCitations []*SourceCitation
}

func parseEvent(ctx context.Context, line *gedcom7.Line, subnodes []*gedcom.Node) (out *Event, err error) {
	out = &Event{}

	var subline *gedcom7.Line

	for _, subnode := range subnodes {
		if subline, err = parseLine(subnode); err != nil {
			return
		}

		fields := map[string]any{
			"func":    "parseEvent",
			"line":    line.Text,
			"subtag":  subline.Tag,
			"subline": subline.Text,
		}

		log.Debug(ctx, fields, "")

		switch subline.Tag {
		case "DATE":
			if out.Date != nil {
				err = fmt.Errorf("error parsing event, multiple DATE lines, conflicting line: %q", subline.Text)
				return
			}

			out.Date, err = newDate(subline.Payload)
			if err != nil {
				return
			}
		case "PLAC":
			if out.Place != "" {
				err = fmt.Errorf("error parsing event, multiple PLAC lines, conflicting line: %q", subline.Text)
				return
			}

			out.Place = subline.Payload
		case "SOUR":
			citation, err := parseSourceCitation(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				return nil, fmt.Errorf("error parsing source citation: %w", err)
			}
			out.SourceCitations = append(out.SourceCitations, citation)
		default:
			log.Warn(ctx, fields, "unsupported Tag")
		}
	}

	return
}
