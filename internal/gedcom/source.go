package gedcom

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"
	"github.com/rafaelespinoza/ged/internal/log"
)

// A SourceCitation supports claims made in a superstructure. Its URI is g7:SOUR.
type SourceCitation struct {
	// Xref is the cross-reference ID of a top-level SourceRecord.
	Xref string
	Page string
	// Data is meant to represent extra info about a source. Its URI is G7:SOUR-DATA.
	// This field is rather free-form, there is no payload.
	Data  map[string]string
	Notes []*Note
}

func parseSourceCitation(ctx context.Context, line *gedcom7.Line, subnodes []*gedcom.Node) (out *SourceCitation, err error) {
	out = &SourceCitation{Xref: line.Payload}

	var subline *gedcom7.Line

	for _, subnode := range subnodes {
		if subline, err = parseLine(subnode); err != nil {
			return
		}

		fields := map[string]any{
			"func":    "parseSourceCitation",
			"line":    line.Text,
			"subtag":  subline.Tag,
			"subline": subline.Text,
		}

		log.Debug(ctx, fields, "")

		switch subline.Tag {
		case "PAGE":
			out.Page = subline.Payload
		case "DATA":
			out.Data = make(map[string]string)
			var dataSubline *gedcom7.Line
			for _, dataSubnode := range subnode.GetSubnodes() {
				if dataSubline, err = parseLine(dataSubnode); err != nil {
					err = fmt.Errorf("could not parse SourceCitation Data: %w", err)
					return
				}
				out.Data[dataSubline.Tag] = dataSubline.Payload
			}
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

const sourceCitationSubfieldDelimiter = ":"

// ParsePage interprets the Page field as a richer struct type.
//
// The reason for this method is that the GEDCOM 7 spec recommends that a PAGE delimits its fields with a
// comma. But the problem is that many of the delimited values will have a comma. The delimiter itself can't
// really be part of the delimited value. Therefore, this method requires an explicit delimiter.
func (s *SourceCitation) ParsePage(fieldDelimiter string) (out SourcePage, err error) {
	if fieldDelimiter == sourceCitationSubfieldDelimiter {
		err = fmt.Errorf("input fieldDelimiter cannot be %q; that is already used for Tuples", sourceCitationSubfieldDelimiter)
		return
	}

	parts := strings.Split(s.Page, fieldDelimiter)
	tuples := make(map[string]string)
	components := make([]string, 0, len(parts))

	for _, part := range parts {
		subparts := strings.Split(part, sourceCitationSubfieldDelimiter)

		switch len(subparts) {
		case 0:
			continue
		case 1:
			components = append(components, strings.TrimSpace(subparts[0]))
		case 2:
			key, val := strings.TrimSpace(subparts[0]), strings.TrimSpace(subparts[1])
			if _, ok := tuples[key]; ok {
				err = fmt.Errorf("duplicated key %q", key)
				return
			}
			tuples[key] = val
		default:
			err = fmt.Errorf("invalid Page line (%q); subparts (%q) can only have max 2 delimiters (%q)", s.Page, subparts, sourceCitationSubfieldDelimiter)
			return
		}
	}

	out = SourcePage{
		Payload:    s.Page,
		Components: slices.Clip(components),
		Tuples:     tuples,
	}

	return
}

type SourcePage struct {
	// Payload is the original line text.
	Payload string
	// Components are individual parts of the original line that sit between
	// delimiters, but do not seem to be a key-value pair.
	Components []string
	// Tuples are similar to Components in that they sit between a delimiter in
	// the original line, except they are interpreted as :-separated key-value
	// pairs.
	Tuples map[string]string
}
