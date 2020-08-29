package main

import (
	"bufio"
	"io"
	"strconv"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

func TryToConvertToUTF8(reader io.Reader) (r io.Reader) {
	br := bufio.NewReader(reader)

	peek, err := br.Peek(4)
	if err != nil {
		return br
	}
	switch {
	case peek[0] == 0xef && peek[1] == 0xbb && peek[2] == 0xbf:
		// UTF-8 with BOM
		return unicode.UTF8BOM.NewDecoder().Reader(br)
	case peek[0] == 0xfe && peek[1] == 0xff:
		// UTF16BE with BOM
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder().Reader(br)
	case peek[0] == 0xff && peek[1] == 0xfe && (peek[2] != 0x00 || peek[3] != 00):
		// UTF16LE with BOM
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Reader(br)
	case peek[0] == 0x00 && peek[1] == 0x00 && peek[2] == 0xfe && peek[3] == 0xff:
		// UTF32BE with BOM
		return utf32.UTF32(utf32.BigEndian, utf32.UseBOM).NewDecoder().Reader(br)
	case peek[0] == 0xff && peek[1] == 0xfe && peek[2] == 0x00 && peek[3] == 0x00:
		// UTF32LE with BOM
		return utf32.UTF32(utf32.LittleEndian, utf32.UseBOM).NewDecoder().Reader(br)
	case peek[0] == 0x00 && strconv.IsPrint(rune(peek[1])) && peek[2] == 0x00 && strconv.IsPrint(rune(peek[3])):
		// UTF16BE w/o BOM
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder().Reader(br)
	case strconv.IsPrint(rune(peek[0])) && peek[1] == 0x00 && strconv.IsPrint(rune(peek[2])) && peek[3] == 0x00:
		// UTF16BE w/o BOM
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder().Reader(br)
	case peek[0] == 0x00 && peek[1] == 0x00 && peek[2] == 0x00 && strconv.IsPrint(rune(peek[3])):
		// UTF32BE w/o BOM
		return utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM).NewDecoder().Reader(br)
	case strconv.IsPrint(rune(peek[0])) && peek[1] == 0x00 && peek[2] == 0x00 && peek[3] == 0x00:
		// UTF32LE w/o BOM
		return utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM).NewDecoder().Reader(br)
	default:
		// UTF-8 as-if
		return br
	}
}
