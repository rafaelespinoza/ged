package gedcom

import (
	"context"
	"fmt"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"
	"github.com/rafaelespinoza/ged/internal/log"
)

// A Note is a catch-all location for info that doesn't really fit within other
// defined structures. The Payload field may contain extra research notes,
// context, or alternative interpretations of other data. Its URI is g7:NOTE.
type Note struct {
	Payload string
	// Lang is the primary language for which the Note is written.
	Lang            string
	SourceCitations []*SourceCitation
}

func parseNote(ctx context.Context, line *gedcom7.Line, subnodes []*gedcom.Node) (out *Note, err error) {
	out = &Note{Payload: line.Payload}

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
		case "CONC":
			// This tag was deprecated in GEDCOM v7. The last GEDCOM version for
			// which it was valid was v5.5.1. This tag would only appear in the
			// data if the library parser is configured to allow deprecated
			// tags. See func gedcom7.WithMaxDeprecatedTags.
			out.Payload += subline.Payload
		case "LANG":
			out.Lang = subline.Payload
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
