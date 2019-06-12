package i18n

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	dp = Printer(language.English)
)

// nolint
func Printer(t language.Tag) *message.Printer {
	return message.NewPrinter(t)
}
