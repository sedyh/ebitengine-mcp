package out

import (
	"crypto/sha256"
	"encoding/hex"
	"unicode/utf8"
)

func Short(str string) string {
	return Trunc(Hash([]byte(str)), 10)
}

func Trunc(str string, length int) string {
	if length <= 0 {
		return ""
	}
	if utf8.RuneCountInString(str) < length {
		return str
	}
	return string([]rune(str)[:length])
}

func Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
