package entity

type Lineage struct {
	Description        string
	Type               LineageType
	GenerationsRemoved int
	CommonAncestors    []Person
}

type LineageType int

const (
	Self LineageType = iota
	Sibling
	Child
	Parent
	AuntUncle
	Cousin
	NieceNephew
)

var lineageNames = []string{
	"self",
	"sibling",
	"child",
	"parent",
	"aunt/uncle",
	"cousin",
	"niece/nephew",
}

func (t LineageType) String() string {
	if t < 0 || int(t) >= len(lineageNames) {
		return ""
	}

	return lineageNames[t]
}
