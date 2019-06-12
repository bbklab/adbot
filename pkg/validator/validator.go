package validator

import "github.com/bbklab/adbot/pkg/utils"

// Validator define a common interface to perform common validation checks
type Validator interface {
	Validate() error
}

var (
	// NormalCharacters pre-define common used char set
	NormalCharacters []rune // most general chars [0-9a-zA-Z.-_]
)

func init() {
	NormalCharacters = append(NormalCharacters, utils.LowerAlpha...)
	NormalCharacters = append(NormalCharacters, utils.UpperAlpha...)
	NormalCharacters = append(NormalCharacters, utils.Digits...)
	NormalCharacters = append(NormalCharacters, []rune(".-_")...)
}

// exported functions
// short hands to call each of validator
//

// String perform string validation check
func String(str string, min, max int, chars []rune) error {
	return newStrValidator(str, min, max, chars).Validate()
}

// Int perform int validation check
func Int(val, min, max int) error {
	return newIntValidator(val, min, max).Validate()
}
