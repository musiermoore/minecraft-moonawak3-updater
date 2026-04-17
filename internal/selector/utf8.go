package selector

import (
	"os"
	"unicode/utf8"
)

func readRuneFromFirstByte(first byte) (rune, error) {
	if first < utf8.RuneSelf {
		return rune(first), nil
	}

	width := utf8Width(first)

	buf := []byte{first}
	for len(buf) < width {
		var b [1]byte
		_, err := os.Stdin.Read(b[:])
		if err != nil {
			return 0, err
		}
		buf = append(buf, b[0])

		if r, size := utf8.DecodeRune(buf); r != utf8.RuneError && size == len(buf) {
			return r, nil
		}
	}

	r, _ := utf8.DecodeRune(buf)
	return r, nil
}

func utf8Width(first byte) int {
	switch {
	case first&0xE0 == 0xC0:
		return 2
	case first&0xF0 == 0xE0:
		return 3
	case first&0xF8 == 0xF0:
		return 4
	default:
		return utf8.UTFMax
	}
}
