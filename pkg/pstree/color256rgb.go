package pstree

import "fmt"

var (
	ansiCode string
)

func color256(cm ColorMap, text *string) {
	ansiCode = fmt.Sprintf("\033[1;38;2;%d;%d;%dm", cm.R, cm.G, cm.B)
	*text = fmt.Sprintf("%s%s%s", ansiCode, *text, AnsiReset)
}

func Color256Black(cs ColorScheme, text *string) {
	color256(cs.Black, text)
}

func Color256BlackBold(cs ColorScheme, text *string) {
	color256(cs.BlackBold, text)
}

func Color256Blue(cs ColorScheme, text *string) {
	color256(cs.Blue, text)
}

func Color256BlueBold(cs ColorScheme, text *string) {
	color256(cs.BlueBold, text)
}

func Color256Cyan(cs ColorScheme, text *string) {
	color256(cs.Cyan, text)
}

func Color256CyanBold(cs ColorScheme, text *string) {
	color256(cs.CyanBold, text)
}

func Color256Green(cs ColorScheme, text *string) {
	color256(cs.Green, text)
}

func Color256GreenBold(cs ColorScheme, text *string) {
	color256(cs.GreenBold, text)
}

func Color256Magenta(cs ColorScheme, text *string) {
	color256(cs.Magenta, text)
}

func Color256Orange(cs ColorScheme, text *string) {
	color256(cs.Orange, text)
}

func Color256OrangeBold(cs ColorScheme, text *string) {
	color256(cs.OrangeBold, text)
}

func Color256MagentaBold(cs ColorScheme, text *string) {
	color256(cs.MagentaBold, text)
}

func Color256Red(cs ColorScheme, text *string) {
	color256(cs.Red, text)
}

func Color256RedBold(cs ColorScheme, text *string) {
	color256(cs.RedBold, text)
}

func Color256White(cs ColorScheme, text *string) {
	color256(cs.White, text)
}

func Color256WhiteBold(cs ColorScheme, text *string) {
	color256(cs.WhiteBold, text)
}

func Color256Yellow(cs ColorScheme, text *string) {
	color256(cs.Yellow, text)
}

func Color256YellowBold(cs ColorScheme, text *string) {
	color256(cs.YellowBold, text)
}
