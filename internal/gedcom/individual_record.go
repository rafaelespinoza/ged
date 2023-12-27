package gedcom

import (
	"context"
	"fmt"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"

	"github.com/rafaelespinoza/ged/internal/gedcom/enumset"
	"github.com/rafaelespinoza/ged/internal/log"
)

// IndividualRecord is a record structure for an individual person. Its URI is
// g7:record-IND.
type IndividualRecord struct {
	Xref              string
	Names             []PersonalName
	Sex               enumset.Sex
	Birth             *Event
	Death             *Event
	FamiliesAsChild   []string // Xref IDs of families where the person is a child.
	FamiliesAsPartner []string // Xref IDs of families where the person is a partner, such as a spouse.
	SourceCitations   []*SourceCitation
	Notes             []*Note
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

		switch subline.Tag {
		case "NAME":
			name, err := parsePersonalName(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				return nil, fmt.Errorf("error parsing personal name: %w", err)
			}
			out.Names = append(out.Names, *name)
		case "BIRT":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing BIRT, skipping")
			} else {
				out.Birth = event
			}
		case "DEAT":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing DEAT, skipping")
			} else {
				out.Death = event
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
