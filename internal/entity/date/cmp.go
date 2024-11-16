package date

// CmpDates compares two *Dates. It returns
//
//	-1 if a < b.
//	 0 if a == b.
//	+1 if a > b.
//
// Date fields compared between a and b are Year, Month, Day. A zero value in
// any of those fields indicates that the date is approximated, thus a more
// approximate date is considered less than a more precise date.
func CmpDates(a, b *Date) int {
	if a.Year < b.Year {
		return -1
	} else if a.Year > b.Year {
		return 1
	}

	if a.Month < b.Month {
		return -1
	} else if a.Month > b.Month {
		return 1
	}

	if a.Day < b.Day {
		return -1
	} else if a.Day > b.Day {
		return 1
	}

	return 0
}

// CmpDateRanges compares two *Ranges. It returns
//
//	-1 if a < b.
//	 0 if a == b.
//	+1 if a > b.
//
// Each Range is compared against the other range using its own non-empty
// components: their Lo and Hi fields. If a Range's Lo and Hi fields are both
// non-empty, then its Lo field is used for the comparison.
func CmpDateRanges(a, b *Range) int {
	var left, right *Date
	if a.Lo != nil {
		left = a.Lo
	} else if a.Hi != nil {
		left = a.Hi
	}
	if b.Lo != nil {
		right = b.Lo
	} else if b.Hi != nil {
		right = b.Hi
	}

	if left != nil && right != nil {
		return CmpDates(left, right)
	} else if right == nil {
		return -1
	} else if left == nil {
		return 1
	}

	return 0
}

// CmpDateToDateRange compares *Date and *Range. It returns
//
//	-1 if a < b.
//	 0 if a == b.
//	+1 if a > b.
func CmpDateToDateRange(a *Date, b *Range) int {
	if a != nil && b == nil {
		return -1
	} else if a == nil && b == nil {
		return 0
	} else if a == nil && b != nil {
		return 1
	}

	if b.Lo != nil {
		return CmpDates(a, b.Lo)
	} else if b.Hi != nil {
		return CmpDates(a, b.Hi)
	}
	return 0
}
