package utils

import (
	"math/rand"
	"strings"
	"unicode"
)

// Truncate truncate given string to maximum given length
func Truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		s = s[:maxLen] + " ..."
	}
	return s
}

// StripSpaces remove all white chars in the given string
func StripSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // if the character is a space, drop it
		}
		return r // else keep it in the string
	}, s)
}

// Deobfuscate decode the `Obfuscated` text
func Deobfuscate(s string) string {
	var clear string
	for i := 0; i < len(s); i++ {
		clear += string(int(s[i]) - 1)
	}
	return clear
}

// Obfuscate encode any string text
func Obfuscate(s string) string {
	var obfuscated string
	for i := 0; i < len(s); i++ {
		obfuscated += string(int(s[i]) + 1)
	}
	return obfuscated
}

// Disrupt disrupt any string text
func Disrupt(s string) string {
	var (
		ss        = []byte(s)
		disrupted = make([]byte, 0, len(ss))
	)

	for {
		n := len(ss)
		if n <= 0 {
			break
		}
		idx := rand.Intn(n)
		disrupted = append(disrupted, ss[idx])
		ss = append(ss[:idx], ss[idx+1:]...)
	}

	return string(disrupted)
}
