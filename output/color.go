package output

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	FormatBold   = "bold"
	FormatItalic = "italic"
	ColorBlack   = "black"
	ColorRed     = "red"
	ColorGreen   = "green"
	ColorYellow  = "yellow"
	ColorBlue    = "blue"
	ColorMagenta = "magenta"
	ColorCyan    = "cyan"
	ColorWhite   = "white"
	ColorGray    = "gray"

	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiItalic  = "\033[3m"
	ansiBlack   = "\033[30m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
	ansiWhite   = "\033[37m"
	ansiGray    = "\033[90m"
)

var ansiColors = map[string]string{
	"reset":      ansiReset,
	FormatBold:   ansiBold,
	FormatItalic: ansiItalic,
	ColorBlack:   ansiBlack,
	ColorRed:     ansiRed,
	ColorGreen:   ansiGreen,
	ColorYellow:  ansiYellow,
	ColorBlue:    ansiBlue,
	ColorMagenta: ansiMagenta,
	ColorCyan:    ansiCyan,
	ColorWhite:   ansiWhite,
	ColorGray:    ansiGray,
}

var ansiCodeRegex = regexp.MustCompile(`\033\[[0-9;]*m`)

func ColorText(text string, snippet string, color string) string {
	if snippet == "" {
		return text
	}

	newColorCode, ok := ansiColors[color]
	if !ok {
		return text
	}

	var result strings.Builder
	searchIdx := 0

	for {
		idx := strings.Index(text[searchIdx:], snippet)
		if idx == -1 {
			result.WriteString(text[searchIdx:])
			break
		}

		absoluteIdx := searchIdx + idx
		currentColor := getActiveColor(text[:absoluteIdx])
		restoreColor := currentColor
		if restoreColor == "" {
			restoreColor = ansiReset
		}

		result.WriteString(text[searchIdx:absoluteIdx])
		result.WriteString(newColorCode)               
		result.WriteString(snippet)                    
		result.WriteString(restoreColor)               
		searchIdx = absoluteIdx + len(snippet)         
	}

	return result.String()
}

func getActiveColor(text string) string {
	matches := ansiCodeRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

func joinArgs(args ...interface{}) string {
	strs := make([]string, len(args))
	for i, arg := range args {
		strs[i] = fmt.Sprint(arg)
	}
	return strings.ReplaceAll(strings.Join(strs, ""), ansiReset, "")
}

func Gray(args ...interface{}) string {
	return ansiGray + joinArgs(args...) + ansiReset
}

func Red(args ...interface{}) string {
	return ansiRed + joinArgs(args...) + ansiReset
}

func Green(args ...interface{}) string {
	return ansiGreen + joinArgs(args...) + ansiReset
}

func Yellow(args ...interface{}) string {
	return ansiYellow + joinArgs(args...) + ansiReset
}

func Blue(args ...interface{}) string {
	return ansiBlue + joinArgs(args...) + ansiReset
}

func Cyan(args ...interface{}) string {
	return ansiCyan + joinArgs(args...) + ansiReset
}

func Magenta(args ...interface{}) string {
	return ansiMagenta + joinArgs(args...) + ansiReset
}

func Bold(args ...interface{}) string {
	return ansiBold + joinArgs(args...) + ansiReset
}

func Italic(args ...interface{}) string {
	return ansiItalic + joinArgs(args...) + ansiReset
}

func FormatReset() string {
	return ansiReset
}
