package validator

import (
	"fmt"
)

func newIntValidator(val, min, max int) Validator {
	return &intValidator{val, min, max}
}

type intValidator struct {
	val int // target int to be checked
	min int
	max int
}

func (v *intValidator) Validate() error {
	if v.val < v.min || v.val > v.max {
		return fmt.Errorf("%d must be numberic between [%d,%d]", v.val, v.min, v.max)
	}
	return nil
}
