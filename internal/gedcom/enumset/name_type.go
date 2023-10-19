package enumset

// NameType is g7:enumset-NAME-TYPE.
type NameType string

const (
	AKA          = NameType("AKA")
	Birth        = NameType("BIRTH")
	Immigrant    = NameType("IMMIGRANT")
	Maiden       = NameType("MAIDEN")
	Married      = NameType("MARRIED")
	Professional = NameType("PROFESSIONAL")
	Other        = NameType("OTHER")
)
