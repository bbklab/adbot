package version

import (
	"strconv"
	"strings"
)

// LessThan checks if a version is less than another version
func LessThan(current, other string) bool {
	return Compare(current, other) == -1
}

// LessThanOrEqualTo checks if a version is less than or equal to another
func LessThanOrEqualTo(current, other string) bool {
	return Compare(current, other) <= 0
}

// GreaterThan checks if a version is greater than another one
func GreaterThan(current, other string) bool {
	return Compare(current, other) == 1
}

// GreaterThanOrEqualTo checks ia version is greater than or equal to another
func GreaterThanOrEqualTo(current, other string) bool {
	return Compare(current, other) >= 0
}

// Equal checks if a version is equal to another
func Equal(current, other string) bool {
	return Compare(current, other) == 0
}

// Compare compares two of provided version
func Compare(current, other string) int {
	var (
		currTab  = cutVersion(current)
		otherTab = cutVersion(other)
	)

	max := len(currTab)
	if len(otherTab) > max {
		max = len(otherTab)
	}
	for i := 0; i < max; i++ {
		var currInt, otherInt int

		if len(currTab) > i {
			currInt = numberic(currTab[i])
		}
		if len(otherTab) > i {
			otherInt = numberic(otherTab[i])
		}
		if currInt > otherInt {
			return 1
		}
		if otherInt > currInt {
			return -1
		}
	}
	return 0
}

func cutVersion(ver string) []string {
	ver = strings.NewReplacer([]string{
		"_", ".",
		"-", ".",
	}...).Replace(ver)
	return strings.Split(ver, ".")
}

func numberic(s string) int {
	s = strings.TrimFunc(s, func(r rune) bool {
		return r < '0' || r > '9'
	})

	n, _ := strconv.Atoi(s)
	return n
}
