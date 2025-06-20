package cli

import (
	"regexp"
	"strings"
)

var (
	regNewline = regexp.MustCompile(`\n+`)
	regTab     = regexp.MustCompile(`\t+`)
	regSpace   = regexp.MustCompile(` +`)
	regColor   = regexp.MustCompile("\x1b\\[[0-9;]*[mG]")
)

const sep = ", "

func Trim(log string) string {
	result := strings.ReplaceAll(log, "\r", "")
	result = regNewline.ReplaceAllString(result, "\n")
	result = regTab.ReplaceAllString(result, " ")
	result = regSpace.ReplaceAllString(result, " ")
	result = regColor.ReplaceAllString(result, "")
	result = strings.ReplaceAll(result, sep, "")

	lines := strings.Split(result, "\n")
	res := make([]string, 0, len(lines))
	for _, line := range lines {
		t := strings.Trim(line, "\n")
		t = strings.Trim(t, " ")
		if t == "" {
			continue
		}
		res = append(res, t)
	}

	return strings.Join(res, sep)
}

func Unwrap(log string) []string {
	if log == "" {
		return []string{}
	}
	return strings.Split(log, sep)
}

func Wrap(logs []string) string {
	if len(logs) == 0 {
		return ""
	}
	return strings.Join(logs, sep)
}
