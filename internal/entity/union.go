package entity

// A Union is a relationship between two people, usually resulting in children.
type Union struct {
	ID        string
	Person1   *Person
	Person2   *Person
	StartDate *Date
	EndDate   *Date
	Children  []*Person
}
