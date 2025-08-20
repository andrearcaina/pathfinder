package ui

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func BadgeDisplay(message, value string) string {
	return BadgeStyle().Render(fmt.Sprintf("%s %s", message, value))
}

func FormatInt(number int, lang language.Tag) string {
	p := message.NewPrinter(lang)
	return p.Sprintf("%d", number)
}

func FormatIntBritishEnglish(number int) string {
	return FormatInt(number, language.BritishEnglish)
}

func FormatIntCanadianFrench(number int) string {
	return FormatInt(number, language.CanadianFrench)
}
