package enumset

type Sex string

const (
	Male       = Sex("M")
	Female     = Sex("F")
	NeitherSex = Sex("X")
	SexUnknown = Sex("U")
)

func NewSex(in string) (out Sex) {
	switch in {
	case "M":
		out = Male
	case "F":
		out = Female
	case "X":
		out = NeitherSex
	default:
		out = SexUnknown
	}

	return
}
