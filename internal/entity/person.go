package entity

import "time"

func NewPerson(name PersonalName, bday, dday *time.Time, parents ...*Person) *Person {
	person := Person{
		Name:      name,
		Birthdate: bday,
		Deathdate: dday,
		Parents:   parents,
	}
	return &person
}

// A Person is an individual that existed, is thought to have existed, or still
// exists in real life.
type Person struct {
	ID        string
	Name      PersonalName
	Birthdate *time.Time
	Deathdate *time.Time
	Parents   []*Person
	Children  []*Person
}
