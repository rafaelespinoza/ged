package gedcom

import (
	"context"
	"fmt"
	"slices"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"

	"github.com/rafaelespinoza/ged/internal/entity/date"
	"github.com/rafaelespinoza/ged/internal/gedcom/enumset"
	"github.com/rafaelespinoza/ged/internal/log"
)

// IndividualRecord is a record structure for an individual person. Its URI is
// g7:record-IND.
type IndividualRecord struct {
	Xref              string
	Names             []PersonalName
	Sex               enumset.Sex
	Birth             []*Event
	Baptism           []*Event
	Christening       []*Event
	Residences        []*Event
	Naturalizations   []*Event
	Death             []*Event
	Burial            []*Event
	Events            []*Event // Other events relevant to a person. Denoted by Type field.
	FamiliesAsChild   []string // Xref IDs of families where the person is a child.
	FamiliesAsPartner []string // Xref IDs of families where the person is a partner, such as a spouse.
	SourceCitations   []*SourceCitation
	Notes             []*Note

	sortedEvents []*Event
}

func parseIndividualRecord(ctx context.Context, i int, line *gedcom7.Line, subnodes []*gedcom.Node) (out *IndividualRecord, err error) {
	out = &IndividualRecord{Xref: line.Xref}

	var subline *gedcom7.Line

	for j, subnode := range subnodes {
		if subline, err = parseLine(subnode); err != nil {
			return
		}

		fields := map[string]any{
			"func":    "parseIndividualRecord",
			"i":       i,
			"j":       j,
			"line":    line.Text,
			"subtag":  subline.Tag,
			"subline": subline.Text,
		}

		log.Debug(ctx, fields, "")

		switch tag := subline.Tag; tag {
		case "NAME":
			name, err := parsePersonalName(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				return nil, fmt.Errorf("error parsing personal name: %w", err)
			}
			out.Names = append(out.Names, *name)
		case "BIRT":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing "+tag+", skipping")
			} else {
				event.setTypeIfEmpty("Birth")
				out.Birth = append(out.Birth, event)
			}
		case "BAPM":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing "+tag+", skipping")
			} else {
				event.setTypeIfEmpty("Baptism")
				out.Baptism = append(out.Baptism, event)
			}
		case "CHR":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing "+tag+", skipping")
			} else {
				event.setTypeIfEmpty("Christening")
				out.Christening = append(out.Christening, event)
			}
		case "RESI":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing "+tag+", skipping")
			} else {
				event.setTypeIfEmpty("Residence")
				out.Residences = append(out.Residences, event)
			}
		case "NATU":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing "+tag+", skipping")
			} else {
				event.setTypeIfEmpty("Naturalization")
				out.Naturalizations = append(out.Naturalizations, event)
			}
		case "EVEN":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing "+tag+", skipping")
			} else {
				event.setTypeIfEmpty("Event")
				out.Events = append(out.Events, event)
			}
		case "DEAT":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing "+tag+", skipping")
			} else {
				event.setTypeIfEmpty("Death")
				out.Death = append(out.Death, event)
			}
		case "BURI":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing "+tag+", skipping")
			} else {
				event.setTypeIfEmpty("Burial")
				out.Burial = append(out.Burial, event)
			}
		case "SEX":
			out.Sex = enumset.NewSex(subline.Payload)
		case "FAMC":
			out.FamiliesAsChild = append(out.FamiliesAsChild, subline.Payload)
		case "FAMS":
			out.FamiliesAsPartner = append(out.FamiliesAsPartner, subline.Payload)
		case "SOUR":
			citation, err := parseSourceCitation(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				return nil, fmt.Errorf("error parsing source citation: %w", err)
			}
			out.SourceCitations = append(out.SourceCitations, citation)
		case "NOTE":
			note, err := parseNote(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				return nil, fmt.Errorf("error parsing note: %w", err)
			}
			out.Notes = append(out.Notes, note)
		default:
			log.Warn(ctx, fields, "unsupported Tag")
		}
	}

	return
}

func (i *IndividualRecord) EventLog() []*Event {
	if i.sortedEvents != nil {
		return i.sortedEvents
	}

	allEvents := [][]*Event{
		i.Birth,
		i.Baptism,
		i.Christening,
		i.Residences,
		i.Naturalizations,
		i.Death,
		i.Burial,
		i.Events,
	}
	var totalCount int
	for _, someEvents := range allEvents {
		totalCount += len(someEvents)
	}

	out := make([]*Event, 0, totalCount)
	for _, someEvents := range allEvents {
		for _, event := range someEvents {
			out = append(out, event)
		}
	}

	slices.SortStableFunc(out, func(left, right *Event) int {
		if left.Date != nil && right.Date != nil {
			return date.CmpDates(left.Date, right.Date)
		}
		if left.DateRange != nil && right.DateRange != nil {
			return date.CmpDateRanges(left.DateRange, right.DateRange)
		}

		if left.Date != nil && right.DateRange != nil {
			return date.CmpDateToDateRange(left.Date, right.DateRange)
		} else if left.Date != nil && right.DateRange == nil {
			return -1
		} else if left.Date == nil && right.DateRange != nil {
			return 1
		}
		return 0
	})

	i.sortedEvents = out
	return i.sortedEvents
}
