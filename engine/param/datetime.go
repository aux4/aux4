package param

import (
	"fmt"
	"strings"
	"time"
)

// resolveDateTimeVariables resolves the date/time function-style helpers:
//
//	date()      -> ISO date            (2006-01-02)
//	time()      -> ISO time            (15:04:05)
//	datetime()  -> ISO date-time       (2006-01-02T15:04:05, local)
//	utc()       -> ISO UTC date-time   (2006-01-02T15:04:05Z)
//	epoch()     -> Unix timestamp int  (seconds)
//
// date(), time(), datetime() and utc() accept an optional format argument using
// moment-style tokens (e.g. date(MMM, DD, YY), time(H:m)). epoch() accepts an
// optional unit: s (default), ms or ns.
//
// A single time.Now() is captured for the whole instruction so that multiple
// helpers in the same command stay consistent.
func resolveDateTimeVariables(instruction string) (string, error) {
	now := time.Now()

	instruction = resolveTimeFunc(instruction, "datetime", now, "YYYY-MM-DDTHH:mm:ss")
	instruction = resolveTimeFunc(instruction, "utc", now.UTC(), "YYYY-MM-DDTHH:mm:ss[Z]")
	instruction = resolveTimeFunc(instruction, "date", now, "YYYY-MM-DD")
	instruction = resolveTimeFunc(instruction, "time", now, "HH:mm:ss")
	instruction = resolveEpochFunc(instruction, now)

	return instruction, nil
}

func resolveTimeFunc(instruction, name string, now time.Time, defaultLayout string) string {
	result, _ := resolveFunction(instruction, `\b`+name+`\(([^)]*)\)`, func(groups []string) (string, error) {
		layout := strings.TrimSpace(groups[0])
		if layout == "" {
			layout = defaultLayout
		}
		return formatMoment(now, layout), nil
	})
	return result
}

func resolveEpochFunc(instruction string, now time.Time) string {
	result, _ := resolveFunction(instruction, `\bepoch\(([^)]*)\)`, func(groups []string) (string, error) {
		switch strings.TrimSpace(groups[0]) {
		case "", "s":
			return fmt.Sprintf("%d", now.Unix()), nil
		case "ms":
			return fmt.Sprintf("%d", now.UnixMilli()), nil
		case "ns":
			return fmt.Sprintf("%d", now.UnixNano()), nil
		default:
			return "epoch(" + groups[0] + ")", nil // unknown unit — leave literal
		}
	})
	return result
}

// momentToken maps a moment-style format token to a formatting function.
// Tokens are matched longest-first so that e.g. YYYY wins over YY and HH over H.
type momentToken struct {
	pattern string
	format  func(time.Time) string
}

var momentTokens = []momentToken{
	{"YYYY", func(t time.Time) string { return fmt.Sprintf("%04d", t.Year()) }},
	{"YY", func(t time.Time) string { return fmt.Sprintf("%02d", t.Year()%100) }},
	{"MMMM", func(t time.Time) string { return t.Month().String() }},
	{"MMM", func(t time.Time) string { return t.Month().String()[:3] }},
	{"MM", func(t time.Time) string { return fmt.Sprintf("%02d", int(t.Month())) }},
	{"M", func(t time.Time) string { return fmt.Sprintf("%d", int(t.Month())) }},
	{"DD", func(t time.Time) string { return fmt.Sprintf("%02d", t.Day()) }},
	{"D", func(t time.Time) string { return fmt.Sprintf("%d", t.Day()) }},
	{"dddd", func(t time.Time) string { return t.Weekday().String() }},
	{"ddd", func(t time.Time) string { return t.Weekday().String()[:3] }},
	{"HH", func(t time.Time) string { return fmt.Sprintf("%02d", t.Hour()) }},
	{"H", func(t time.Time) string { return fmt.Sprintf("%d", t.Hour()) }},
	{"hh", func(t time.Time) string { return fmt.Sprintf("%02d", hour12(t)) }},
	{"h", func(t time.Time) string { return fmt.Sprintf("%d", hour12(t)) }},
	{"mm", func(t time.Time) string { return fmt.Sprintf("%02d", t.Minute()) }},
	{"m", func(t time.Time) string { return fmt.Sprintf("%d", t.Minute()) }},
	{"ss", func(t time.Time) string { return fmt.Sprintf("%02d", t.Second()) }},
	{"s", func(t time.Time) string { return fmt.Sprintf("%d", t.Second()) }},
	{"A", func(t time.Time) string { return meridiem(t, true) }},
	{"a", func(t time.Time) string { return meridiem(t, false) }},
	{"ZZ", func(t time.Time) string { return offset(t, false) }},
	{"Z", func(t time.Time) string { return offset(t, true) }},
}

// formatMoment renders t according to a moment-style layout. Unknown characters
// are emitted verbatim, so separators like "-", ":", "/", "T" pass through.
// Text wrapped in square brackets is treated as a literal (e.g. [Z] emits "Z"),
// which is how token letters can be included without being interpreted.
func formatMoment(t time.Time, layout string) string {
	var b strings.Builder
	runes := []rune(layout)
	for i := 0; i < len(runes); {
		if runes[i] == '[' {
			if end := indexRune(runes, ']', i+1); end != -1 {
				b.WriteString(string(runes[i+1 : end]))
				i = end + 1
				continue
			}
		}

		matched := false
		for _, token := range momentTokens {
			n := len(token.pattern)
			if i+n <= len(runes) && string(runes[i:i+n]) == token.pattern {
				b.WriteString(token.format(t))
				i += n
				matched = true
				break
			}
		}
		if !matched {
			b.WriteRune(runes[i])
			i++
		}
	}
	return b.String()
}

func indexRune(runes []rune, target rune, from int) int {
	for i := from; i < len(runes); i++ {
		if runes[i] == target {
			return i
		}
	}
	return -1
}

func hour12(t time.Time) int {
	h := t.Hour() % 12
	if h == 0 {
		return 12
	}
	return h
}

func meridiem(t time.Time, upper bool) string {
	if t.Hour() < 12 {
		if upper {
			return "AM"
		}
		return "am"
	}
	if upper {
		return "PM"
	}
	return "pm"
}

func offset(t time.Time, colon bool) string {
	_, seconds := t.Zone()
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if colon {
		return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
	}
	return fmt.Sprintf("%s%02d%02d", sign, hours, minutes)
}
