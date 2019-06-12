package api

import (
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/bbklab/adbot/i18n"
	"github.com/bbklab/adbot/pkg/httpmux"
)

var (
	defaultLang = language.English
)

func getLang(ctx *httpmux.Context) language.Tag {
	if ctx == nil {
		return defaultLang
	}

	var al = ctx.Req.Header.Get("Accept-Language")
	tt, _, err := language.ParseAcceptLanguage(al)
	if err != nil {
		return defaultLang
	}

	if len(tt) == 0 {
		return defaultLang
	}

	if strings.HasPrefix(strings.ToLower(tt[0].String()), "zh") {
		return language.Chinese
	}

	return defaultLang
}

func i18nPrinter(ctx *httpmux.Context) *message.Printer {
	return i18n.Printer(getLang(ctx))
}
