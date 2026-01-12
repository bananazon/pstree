package pstree

import (
	"fmt"
	"strings"
)

func color8(cm ColorMap, text *string) {
	*text = fmt.Sprintf("%s%s%s", cm.Ansi, *text, AnsiReset)
}

func Color8Black(cs ColorScheme, text *string) {
	color8(cs.Black, text)
}

func Color8BlackBold(cs ColorScheme, text *string) {
	color8(cs.BlackBold, text)
}

func Color8Blue(cs ColorScheme, text *string) {
	color8(cs.Blue, text)
}

func Color8BlueBold(cs ColorScheme, text *string) {
	color8(cs.BlueBold, text)
}

func Color8Cyan(cs ColorScheme, text *string) {
	color8(cs.Cyan, text)
}

func Color8CyanBold(cs ColorScheme, text *string) {
	color8(cs.CyanBold, text)
}

func Color8Green(cs ColorScheme, text *string) {
	color8(cs.Green, text)
}

func Color8GreenBold(cs ColorScheme, text *string) {
	color8(cs.GreenBold, text)
}

func Color8Magenta(cs ColorScheme, text *string) {
	color8(cs.Magenta, text)
}

func Color8MagentaBold(cs ColorScheme, text *string) {
	color8(cs.MagentaBold, text)
}

func Color8Red(cs ColorScheme, text *string) {
	color8(cs.Red, text)
}

func Color8RedBold(cs ColorScheme, text *string) {
	color8(cs.RedBold, text)
}

func Color8Yellow(cs ColorScheme, text *string) {
	color8(cs.Yellow, text)
}

func Color8YellowBold(cs ColorScheme, text *string) {
	color8(cs.YellowBold, text)
}

func Color8White(cs ColorScheme, text *string) {
	color8(cs.White, text)
}

func Color8WhiteBold(cs ColorScheme, text *string) {
	color8(cs.WhiteBold, text)
}

func Print8ColorRainbow(text string) string {
	cs := ColorSchemes["ansi8"]

	// Color functions that modify a string pointer
	colorFuncs := []func(cs ColorScheme, text *string){
		Color8RedBold,
		Color8YellowBold,
		Color8GreenBold,
		Color8CyanBold,
		Color8BlueBold,
		Color8MagentaBold,
	}

	var builder strings.Builder
	var result string

	for i, ch := range text {
		s := string(ch)
		colorFunc := colorFuncs[i%len(colorFuncs)]
		colorFunc(cs, &s)
		result += s
	}

	builder.WriteString(result)
	return builder.String()
}
