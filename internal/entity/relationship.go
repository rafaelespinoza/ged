package entity

// MutualRelationship describes a relationship from person A to person B and its
// complementary relationship from person B to person A.
//
// If person A and person B have a blood relationship, then CommonPerson is a
// shared common ancestor. The Path fields in each Relationship denote the
// ancestral path to the CommonPerson, where the first person in the Path is
// starting person and the last person is the CommonPerson.
//
// If person A and person B are related by marriage, then Union is non-empty.
// The Path fields in each Relationship describe how the starting person relates
// to a person in the Union.
type MutualRelationship struct {
	CommonPerson *Person
	Union        *Union
	R1, R2       Relationship
}

// Relationship is a unidirectional descriptor how a person at SourceID relates
// to a person at TargetID. The connection may have a consanguineous (by blood)
// nature, or may be affinal (by law, or via marriage).
type Relationship struct {
	// SourceID identifies the person representing the vantage point in this Relationship.
	SourceID string
	// TargetID identifies the person who is being related to by the person at SourceID.
	TargetID string
	// Type may indicate if the relationship is consanguineous (by blood),
	// affinal (by law, or via marriage), or is unknown.
	Type RelationshipType
	// Description elaborates on the Type.
	Description        string
	GenerationsRemoved int
	// Path is the path to a common person if the Type field indicates a
	// consanguineous relationship. If the Type indiciates an affinal
	// relationship, then Path is the path from the person at SourceID to a
	// person in a Union.
	Path []Person
}

type RelationshipType int

const (
	Unknown RelationshipType = iota

	// These values indicate a consanguineous (by blood) relationship.
	Self
	Sibling
	Child
	Parent
	AuntUncle
	Cousin
	NieceNephew

	// These values indicate an affinal (by marriage) relationship.
	Spouse
	SiblingInLaw
	ChildInLaw
	ParentInLaw
	AuntUncleInLaw
	CousinInLaw
	NieceNephewInLaw
)

var relationshipNames = []string{
	"unknown",

	"self",
	"sibling",
	"child",
	"parent",
	"aunt/uncle",
	"cousin",
	"niece/nephew",

	"spouse",
	"sibling in-law",
	"child in-law",
	"parent in-law",
	"aunt/uncle in-law",
	"cousin in-law",
	"niece/nephew in-law",
}

func (t RelationshipType) String() string {
	if t < 0 || int(t) >= len(relationshipNames) {
		return ""
	}

	return relationshipNames[t]
}
