package whisper

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Sanitize cleans model output into a whisper line, or "" if unusable.
// Handles GLM-family reasoning output ( conspiracy...</think> blocks and a
// separate reasoning field are dropped upstream; quote/markdown wrappers,
// filler prefixes and anything over ~80 runes are stripped here.
func Sanitize(s string) string {
	// Drop complete reasoning blocks, then any dangling unclosed tag.
	// "<think" covers both <think> and <thinking>; "<spiracy" covers
	// conspiracy-style dumps.
	for _, tag := range [][2]string{{"<think", "</think"}, {"<spiracy", "</spiracy"}} {
		s = stripPaired(s, tag[0], tag[1])
		if i := strings.Index(strings.ToLower(s), tag[0]); i >= 0 {
			s = s[:i]
		}
	}

	// Keep the first line that actually says something.
	for _, line := range splitLines(s) {
		line = cleanOne(line)
		n := utf8.RuneCountInString(line)
		if n > 80 || n < 2 {
			continue
		}
		return line
	}
	return ""
}

func stripPaired(s, open, close string) string {
	low := strings.ToLower(s)
	for {
		i := strings.Index(low, open)
		if i < 0 {
			return s
		}
		j := strings.Index(low[i:], close)
		if j < 0 {
			return s[:i]
		}
		s = s[:i] + s[i+j+len(close):]
		low = strings.ToLower(s)
	}
}

func cleanOne(line string) string {
	const wrappers = " \t\r\n\"'`“”‘’*«»<>-—.: "
	line = strings.Trim(line, wrappers)
	low := strings.ToLower(line)
	// Strip common assistant filler prefixes.
	for _, pre := range []string{"sure — ", "sure - ", "okay, ", "here's one: ", "here is one: ", "here it is: "} {
		if strings.HasPrefix(low, pre) {
			line = line[len(pre):]
			low = strings.ToLower(line)
		}
	}
	line = strings.Trim(line, wrappers+"-—.: ")
	// Reject replacement chars and control characters (tab is whitespace
	// and handled by the collapse pass below).
	for _, r := range line {
		if r == utf8.RuneError || (r < 32 && r != '\t' && r != 0) {
			return ""
		}
	}
	// Collapse inner whitespace runs to single spaces.
	var b strings.Builder
	prevSpace := false
	for _, r := range line {
		if unicode.IsSpace(r) {
			prevSpace = true
			continue
		}
		if prevSpace && b.Len() > 0 {
			b.WriteByte(' ')
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}
