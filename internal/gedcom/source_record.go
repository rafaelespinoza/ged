package gedcom

import (
	"context"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"
	"github.com/rafaelespinoza/ged/internal/log"
)

// SourceRecord is a record structure for a source. Its URI g7:record-SOUR.
type SourceRecord struct {
	Xref          string
	Title         string
	Author        string
	Abbreviation  string
	Publication   string
	Text          string
	RepositoryIDs []string
	Note          string
	OriginalLines []string

	// TODO: add other fields such as Data, MultimediaLink, ChangeDate, CreationDate, as needed.
}

func parseSourceRecord(ctx context.Context, i int, line *gedcom7.Line, subnodes []*gedcom.Node) (out *SourceRecord, err error) {
	out = &SourceRecord{Xref: line.Xref}

	var subline *gedcom7.Line

	for j, subnode := range subnodes {
		if subline, err = parseLine(subnode); err != nil {
			return
		}

		fields := map[string]any{
			"func":    "parseSourceRecord",
			"i":       i,
			"j":       j,
			"line":    line.Text,
			"subtag":  subline.Tag,
			"subline": subline.Text,
		}

		log.Debug(ctx, fields, "")

		switch subline.Tag {
		case "TITL":
			out.Title = subline.Payload
		case "AUTH":
			out.Author = subline.Payload
		case "ABBR":
			out.Abbreviation = subline.Payload
		case "PUBL":
			out.Publication = subline.Payload
		case "TEXT":
			out.Text = subline.Payload
		case "REPO":
			out.RepositoryIDs = append(out.RepositoryIDs, subline.Payload)
		case "NOTE":
			out.Note = subline.Payload
		default:
			log.Warn(ctx, fields, "unsupported Tag")
		}
	}

	return
}
