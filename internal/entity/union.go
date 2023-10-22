package entity

import "time"

// A Union is a relationship between two people, usually resulting in children.
type Union struct {
	ID        string
	Person1   *Person
	Person2   *Person
	StartDate *time.Time
	EndDate   *time.Time
	Children  []*Person
}
