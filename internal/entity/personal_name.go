package entity

import "strings"

type PersonalName struct {
	Forename     string
	Middle       string
	Surname      string
	BirthSurname string
	Nickname     string
	Suffix       string

	fullname string
}

func (n *PersonalName) Full() string {
	// A potential edge case that may lead to more than 1 invocation of this
	// entire method body, is if all the input name fields are empty. If this
	// becomes a problem, consider changing the type of the unexported field,
	// fullname, to a *string, so that the result of the first invocation may be
	// signaled to subsequent invocations. If the name fields are indeed, all
	// empty, then a second invocation would return a non-nil pointer to an
	// empty string, which is different than a nil pointer.
	if n.fullname != "" {
		return n.fullname
	}
	var nickname string
	if n.Nickname != "" {
		nickname = `"` + n.Nickname + `"`
	}

	allParts := []string{n.Forename, nickname, n.Middle, n.BirthSurname, n.Surname, n.Suffix}
	nonEmptyParts := make([]string, 0, len(allParts))
	for _, name := range allParts {
		if name != "" {
			nonEmptyParts = append(nonEmptyParts, name)
		}
	}
	n.fullname = strings.Join(nonEmptyParts, " ")
	return n.fullname
}
