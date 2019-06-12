package utils

import (
	"errors"
)

// GenPassword generate password with specified size length
// contains: lowers, uppers, digits, specials
func GenPassword(size int) (string, error) {
	return genPassword(size, true)
}

// GenPasswordNonSpecial similar as above but without special chars
func GenPasswordNonSpecial(size int) (string, error) {
	return genPassword(size, false)
}

func genPassword(size int, withSpecial bool) (string, error) {
	if size < 4 {
		return "", errors.New("password length should >= 4")
	}

	// cut the length into 4 pieces
	var (
		n = 4
		x = size / n
		y = size % n
	)

	sets := make([]int, 0, 4)
	for i := 1; i <= n; i++ {
		if i == n { // make the last item with the final rest
			sets = append(sets, x+y)
		} else {
			sets = append(sets, x)
		}
	}

	// generate each length of specified chars
	var (
		uppers   = RandomStringRange(sets[0], UpperAlpha)
		lowers   = RandomStringRange(sets[1], LowerAlpha)
		digits   = RandomStringRange(sets[2], Digits)
		specials string
	)

	if withSpecial {
		specials = RandomStringRange(sets[3], Specials)
	} else {
		specials = RandomStringRange(sets[3], Digits)
	}

	return Disrupt(uppers + lowers + digits + specials), nil
}
