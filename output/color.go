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
	if snippet == "" || !ColorEnabled() {
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

// colorize wraps the joined arguments in an ANSI code, unless the resolved
// color policy says aux4 should not emit color (see colorpolicy.go), in which
// case the plain text is returned.
func colorize(ansiCode string, args ...interface{}) string {
	text := joinArgs(args...)
	if !ColorEnabled() {
		return text
	}
	return ansiCode + text + ansiReset
}

func Gray(args ...interface{}) string {
	return colorize(ansiGray, args...)
}

func Red(args ...interface{}) string {
	return colorize(ansiRed, args...)
}

func Green(args ...interface{}) string {
	return colorize(ansiGreen, args...)
}

func Yellow(args ...interface{}) string {
	return colorize(ansiYellow, args...)
}

func Blue(args ...interface{}) string {
	return colorize(ansiBlue, args...)
}

func Cyan(args ...interface{}) string {
	return colorize(ansiCyan, args...)
}

func Magenta(args ...interface{}) string {
	return colorize(ansiMagenta, args...)
}

func Bold(args ...interface{}) string {
	return colorize(ansiBold, args...)
}

func Italic(args ...interface{}) string {
	return colorize(ansiItalic, args...)
}

func FormatReset() string {
	if !ColorEnabled() {
		return ""
	}
	return ansiReset
}
