package entity

// A Person is an individual that existed, is thought to have existed, or still
// exists in real life.
type Person struct {
	ID        string
	Name      PersonalName
	Birthdate *Date
	Deathdate *Date
	Parents   []*Person
	Children  []*Person
	Spouses   []*Person
}
