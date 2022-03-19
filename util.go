package main

import (
	"bytes"
	"unicode/utf8"
)

func removeNonUTF8Bytes(buf []byte) []byte {
	fn := func(r rune) rune {
		if r == utf8.RuneError {
			return -1
		}
		return r
	}
	return bytes.Map(fn, buf)
}
