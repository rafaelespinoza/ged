package gedcom

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"

	"github.com/rafaelespinoza/logg"
)

// FamilyRecord is a record structure for a family. Its URI g7:record-FAM.
type FamilyRecord struct {
	Xref            string
	ParentXrefs     []string
	ChildXrefs      []string
	MarriedAt       *Event
	DivorcedAt      *Event
	AnnulledAt      *Event
	SourceCitations []*SourceCitation
	Notes           []*Note
}

func parseFamilyRecord(ctx context.Context, i int, line *gedcom7.Line, subnodes []*gedcom.Node) (out *FamilyRecord, err error) {
	out = &FamilyRecord{Xref: line.Xref}

	var subline *gedcom7.Line

	logger := logg.New("", slog.String("func", "parseFamilyRecord"), slog.Int("i", i))

	for j, subnode := range subnodes {
		if subline, err = parseLine(subnode); err != nil {
			return
		}

		logger.Debug("reading line",
			slog.Int("j", j),
			slog.String("line", line.Text), slog.String("subtag", subline.Tag), slog.String("subline", subline.Text),
		)

		switch tag := subline.Tag; tag {
		case "HUSB", "WIFE":
			out.ParentXrefs = append(out.ParentXrefs, subline.Payload)
		case "CHIL":
			out.ChildXrefs = append(out.ChildXrefs, subline.Payload)
		case "MARR":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				logger.Warn("error parsing tag, skipping", slog.String("tag", tag), slog.Any("error", err))
			} else {
				out.MarriedAt = event
			}
		case "DIV":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				logger.Warn("error parsing tag, skipping", slog.String("tag", tag), slog.Any("error", err))
			} else {
				out.DivorcedAt = event
			}
		case "ANUL":
			event, err := parseEvent(ctx, subline, subnode.GetSubnodes())
			if err != nil {
				logger.Warn("error parsing tag, skipping", slog.String("tag", tag), slog.Any("error", err))
			} else {
				out.AnnulledAt = event
			}
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
			logger.Warn("unsupported Tag, skipping", slog.String("tag", tag))
		}
	}

	return
}
