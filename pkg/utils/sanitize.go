// Original code from https://github.com/subosito/gozaru
//
// Copyright (c) 2014 Alif Rachmawadi <subosito@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//
// Modifications Copyright (c) 2026 cxntered
// - Removed unnecessary dependencies

package utils

import (
	"unicode"
)

var (
	FallbackFilename     = "file"
	WindowsReservedNames = [...]string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5",
		"COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5",
		"LPT6", "LPT7", "LPT8", "LPT9",
	}
)

func Sanitize(s string) string {
	return sanitize(s, 0, "")
}

func SanitizeFallback(s string, fallback string) string {
	return sanitize(s, 0, fallback)
}

func SanitizePad(s string, n int) string {
	return sanitize(s, n, "")
}

func SanitizePadFallback(s string, n int, fallback string) string {
	return sanitize(s, n, fallback)
}

func sanitize(s string, n int, fallback string) string {
	if fallback == "" {
		fallback = FallbackFilename
	}

	sc := clean(s, fallback)
	nc := len(sc)

	if n > nc {
		return sc
	}

	if nc > 255 {
		nc = 255
	}

	if n != 0 {
		nc -= n
	}

	return sc[0:nc]
}

func clean(s string, fallback string) string {
	return filter(normalize(s), fallback)
}

func filter(s string, fallback string) string {
	s = filterWindowsReservedNames(s, fallback)
	s = filterBlank(s, fallback)
	s = filterDot(s, fallback)

	return s
}

func filterWindowsReservedNames(s string, fallback string) string {
	us := toUpper(s)

	for i := range WindowsReservedNames {
		v := WindowsReservedNames[i]

		if v == us {
			return fallback
		}
	}

	return s
}

func filterBlank(s string, fallback string) string {
	if s == "" {
		return fallback
	}

	return s
}

func filterDot(s string, fallback string) string {
	if hasPrefix(s, ".") {
		return fallback + s
	}

	return s
}

func toUpper(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		runes[i] = unicode.ToUpper(r)
	}

	return string(runes)
}

func hasPrefix(s string, prefix string) bool {
	runes := []rune(s)
	prunes := []rune(prefix)
	if len(prunes) > len(runes) {
		return false
	}

	for i, r := range prunes {
		if runes[i] != r {
			return false
		}
	}

	return true
}

func normalize(s string) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes))
	lastWasSpace := false

	for _, r := range runes {
		if isDisallowedRune(r) {
			continue
		}
		if unicode.IsSpace(r) {
			if len(out) == 0 || lastWasSpace {
				continue
			}
			out = append(out, ' ')
			lastWasSpace = true
			continue
		}

		out = append(out, r)
		lastWasSpace = false
	}

	if lastWasSpace && len(out) > 0 {
		out = out[:len(out)-1]
	}

	return string(out)
}

func isDisallowedRune(r rune) bool {
	if r <= 0x1F {
		return true
	}

	switch r {
	case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
		return true
	default:
		return false
	}
}
