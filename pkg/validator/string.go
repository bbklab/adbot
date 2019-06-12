package validator

import (
	"errors"
	"fmt"
	"strings"
)

func newStrValidator(str string, min, max int, chars []rune) Validator {
	return &strValidator{str, min, max, chars}
}

type strValidator struct {
	str   string // target string to be checked
	min   int    // 0/negative means unlimited
	max   int    // 0/negative means unlimited
	chars []rune // empty means unlimited
}

func (v *strValidator) Validate() error {
	if l, n := v.min, len(v.str); l > 0 && n < l {
		if n == 0 {
			return errors.New("required, can't be empty")
		}
		return fmt.Errorf("length can't smaller than %d", l)
	}

	if l, n := v.max, len(v.str); l > 0 && n > l {
		return fmt.Errorf("length can't larger than %d", l)
	}

	if len(v.chars) > 0 {
		for _, r := range v.str {
			if !strings.ContainsRune(string(v.chars), r) {
				return fmt.Errorf("character %q not allowed", r)
			}
		}
	}

	return nil
}
