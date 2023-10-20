package gedcom

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"

	"github.com/rafaelespinoza/ged/internal/log"
)

// Records is a collection of top-level record types.
type Records struct {
	Individuals []*IndividualRecord
	Families    []*FamilyRecord
}

// ReadRecords reads constructs Records out of the input document r.
func ReadRecords(ctx context.Context, r io.Reader) (*Records, error) {
	doc := gedcom7.NewDocument(bufio.NewScanner(r), gedcom7.WithMaxDeprecatedTags("5.5.1"))

	warnings := doc.GetWarnings()
	fields := map[string]any{
		"func":         "ReadRecords",
		"num_records":  doc.Len(),
		"num_warnings": len(warnings),
		"warnings":     warnings,
	}

	log.Info(ctx, fields, "processed gedcom7 document")

	nodes := doc.Records()
	out := Records{
		Individuals: make([]*IndividualRecord, 0, len(nodes)),
		Families:    make([]*FamilyRecord, 0, len(nodes)),
	}

	for i, node := range nodes {
		line, err := parseLine(node)
		if err != nil {
			return nil, fmt.Errorf("item[%d], %w", i, err)
		}

		switch line.Tag {
		case "INDI":
			individual, err := parseIndividualRecord(ctx, i, line, node.GetSubnodes())
			if err != nil {
				return nil, fmt.Errorf("error parsing individual record, line=%q: %w", line.String(), err)
			}
			out.Individuals = append(out.Individuals, individual)
		case "FAM":
			family, err := parseFamilyRecord(ctx, i, line, node.GetSubnodes())
			if err != nil {
				return nil, fmt.Errorf("error parsing family record, line=%q: %w", line.String(), err)
			}
			out.Families = append(out.Families, family)
		default:
			fields := map[string]any{
				"func": "newRecords",
				"i":    i,
				"line": line.Text,
				"tag":  line.Tag,
			}
			log.Warn(ctx, fields, "unsupported Tag")
		}
	}

	return &out, nil
}

func parseLine(node *gedcom.Node) (line *gedcom7.Line, err error) {
	switch val := node.GetValue().(type) {
	case *gedcom7.Line:
		line = val
	default:
		return nil, fmt.Errorf("unsupported type, %T", val)
	}
	return
}
