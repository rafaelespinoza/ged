package gedcom

import (
	"regexp"
	"strings"

	"github.com/funwithbots/go-gedcom/pkg/gedcom"
	"github.com/funwithbots/go-gedcom/pkg/gedcom7"

	"github.com/rafaelespinoza/reltree/internal/gedcom/enumset"
)

// PersonalName is an individual's name. Its URI is g7:INDI-NAME.
type PersonalName struct {
	Payload string
	Type    enumset.NameType

	// The following fields are PERSONAL_NAME_PIECES.
	NamePrefix    string // URI is g7:NPFX
	Given         string // URI is g7:GIVN
	Nickname      string // URI is g7:NICK
	SurnamePrefix string // URI is g7:SPFX
	Surname       string // URI is g7:SURN
	NameSuffix    string // URI is g7:NSFX
}

var surnamePattern = regexp.MustCompile(`(\/[A-z|\s]*\/)`)

func newPersonalName(line *gedcom7.Line, subnodes []*gedcom.Node) (out *PersonalName, err error) {
	out = &PersonalName{Payload: line.Payload}

	var subline *gedcom7.Line

	for _, subnode := range subnodes {
		if subline, err = parseLine(subnode); err != nil {
			return
		}

		payload := subline.Payload

		switch subline.Tag {
		case "NPFX":
			out.NamePrefix = payload
		case "GIVN":
			out.Given = payload
		case "NICK":
			out.Nickname = payload
		case "SPFX":
			out.SurnamePrefix = payload
		case "SURN":
			out.Surname = payload
		case "NSFX":
			out.NameSuffix = payload
		default:
			// There may be some metadata-related tags such as NAME-TYPE, SOUR
			// (Source), or NOTE. For now, not parsing those, but might try to
			// do so later.
		}
	}

	if out.Surname == "" && out.Given == "" {
		input := line.Payload
		var surnameIndex int
		surnameIndex = len(input)

		match := surnamePattern.FindStringSubmatch(input)
		if match != nil {
			out.Surname = strings.Trim(match[1], "/")
			indexMatch := surnamePattern.FindStringSubmatchIndex(input)
			surnameIndex = indexMatch[2]

			input = strings.Replace(input, match[1], "", 1)
		}

		out.Given = strings.TrimSpace(input[:surnameIndex])

		// See if there's a suffix. TODO: clean up this code!
		parts := strings.Fields(input)
		if len(parts) > 1 && !strings.Contains(out.Given, parts[len(parts)-1]) {
			out.NameSuffix = parts[len(parts)-1]
		}
	}

	return
}
