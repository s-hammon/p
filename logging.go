package p

func Colorize(color int, text string) string {
	return Format("\x1b[%dm%v\x1b[0m", color, text)
}
