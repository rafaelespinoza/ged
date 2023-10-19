package gedcom

import (
	"context"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"

	"github.com/rafaelespinoza/reltree/internal/log"
)

// FamilyRecord is a record structure for a family. Its URI g7:record-FAM.
type FamilyRecord struct {
	Xref        string
	ParentXrefs []string
	ChildXrefs  []string
	MarriedAt   *Event
	DivorcedAt  *Event
	AnnulledAt  *Event
}

func parseFamilyRecord(ctx context.Context, i int, line *gedcom7.Line, subnodes []*gedcom.Node) (out *FamilyRecord, err error) {
	out = &FamilyRecord{Xref: line.Xref}

	var subline *gedcom7.Line

	for j, subnode := range subnodes {
		if subline, err = parseLine(subnode); err != nil {
			return
		}

		fields := map[string]any{
			"func":    "parseFamilyRecord",
			"i":       i,
			"j":       j,
			"line":    line.Text,
			"subtag":  subline.Tag,
			"subline": subline.Text,
		}

		log.Debug(ctx, fields, "")

		switch subline.Tag {
		case "HUSB", "WIFE":
			out.ParentXrefs = append(out.ParentXrefs, subline.Payload)
		case "CHIL":
			out.ChildXrefs = append(out.ChildXrefs, subline.Payload)
		case "MARR":
			event, err := newEvent(subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing MARR, skipping")
			} else {
				out.MarriedAt = event
			}
		case "DIV":
			event, err := newEvent(subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing DIV, skipping")
			} else {
				out.DivorcedAt = event
			}
		case "ANUL":
			event, err := newEvent(subline, subnode.GetSubnodes())
			if err != nil {
				log.Error(ctx, fields, err, "error parsing ANUL, skipping")
			} else {
				out.AnnulledAt = event
			}
		default:
			log.Warn(ctx, fields, "unsupported Tag")
		}
	}

	return
}
