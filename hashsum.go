package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

type BadLineError struct {
	Path string
	Line int
	Name string
}

func (e BadLineError) Error() string {
	return fmt.Sprintf("%s: %d: improperly formatted %s checksum line", e.Path, e.Line, e.Name)
}

type HashSumReadFunc func(sum []byte, name string, err error)

type HashSumReader struct {
	Name  string // hash name used in tag
	Width int    // hash result bit width
	Tag   bool   // bsd tag format, or gnu format
	Zero  bool   // '\0' for line separation or not
	CrLf  bool   // lines can ending with CRLF
}

func (r HashSumReader) Read(path string, readFn HashSumReadFunc) {
	file, err := OpenFile(path)
	if err != nil {
		readFn(nil, path, err)
		return
	}
	defer file.Close()

	scn := bufio.NewScanner(TryToConvertToUTF8(file))
	if r.Zero {
		scn.Split(byteTerminatedScanner('\x00'))
	} else if !r.CrLf {
		scn.Split(byteTerminatedScanner('\n'))
	}

	lineParser := r.parseGnuSum
	if r.Tag {
		lineParser = r.parseBsdSum
	}

	var validLineCount uint64
	var lineNo int
	for scn.Scan() {
		lineNo++
		sum, file, ok := lineParser(scn.Text())
		if ok && file == "-" && path == "-" {
			ok = false
		}
		if ok {
			readFn(sum, file, nil)
			validLineCount++
		} else {
			readFn(nil, path, BadLineError{Path: path, Line: lineNo, Name: r.Name})
		}
	}
	if err := scn.Err(); err != nil {
		readFn(nil, path, err)
	} else if validLineCount == 0 {
		readFn(nil, path, fmt.Errorf("%s: no properly formatted %s checksum lines found", path, r.Name))
	}
}

func byteTerminatedScanner(b byte) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, b); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
}

func (r HashSumReader) parseGnuSum(line string) (sum []byte, file string, ok bool) {
	hexWidth := r.Width * 2
	escaped := strings.HasPrefix(line, "/")
	if escaped {
		line = line[1:]
	}
	if len(line) <= hexWidth+2 {
		return
	}
	if line[hexWidth] != ' ' || (line[hexWidth+1] != ' ' && line[hexWidth+1] != '*') {
		return
	}
	var hexSum string
	hexSum, file, ok = line[:hexWidth], line[hexWidth+2:], true
	if escaped {
		file = unescapeName(file)
	}
	sum, err := hex.DecodeString(hexSum)
	if err != nil {
		ok = false
		return
	}
	return
}

func (r HashSumReader) parseBsdSum(line string) (sum []byte, file string, ok bool) {
	hexWidth := r.Width * 2
	escaped := strings.HasPrefix(line, "/")
	if escaped {
		line = line[1:]
	}
	if len(line) <= hexWidth+len(r.Name)+6 {
		return
	}
	if line[:len(r.Name)] != r.Name &&
		line[len(r.Name):len(r.Name)+2] != "( " &&
		line[len(line)-hexWidth-4:len(line)-hexWidth] != ") = " {
		return
	}
	var hexSum string
	hexSum, file, ok = line[len(line)-hexWidth:], line[len(r.Name)+2:len(line)-hexWidth-4], true
	if escaped {
		file = unescapeName(file)
	}
	sum, err := hex.DecodeString(hexSum)
	if err != nil {
		ok = false
		return
	}
	return
}

type HashSumWriter struct {
	Name   string // hash name used in tag
	Tag    bool   // bsd tag format or gnu format
	Zero   bool   // '\0' as line separator or not
	Binary bool   // use '*' in gnu format or not
}

func (w HashSumWriter) Write(out io.Writer, sum []byte, name string) {
	sep := "\n"
	if w.Zero {
		sep = "\x00"
	}
	prefix := ""
	if !w.Zero {
		var escaped bool
		name, escaped = escapeName(name)
		if escaped {
			prefix = `\`
		}
	}
	if w.Tag {
		_, _ = fmt.Fprintf(out, "%s%s (%s) = %s%s", prefix, w.Name, name, hex.EncodeToString(sum), sep)
	} else {
		flag := ' '
		if w.Binary {
			flag = '*'
		}
		_, _ = fmt.Fprintf(out, "%s%s %c%s%s", prefix, hex.EncodeToString(sum), flag, name, sep)
	}
}

func unescapeName(name string) string {
	var sb strings.Builder
	sb.Grow(len(name))
	var escaping bool
	for _, ch := range name {
		if escaping {
			escaping = false
			switch ch {
			case '\\':
				sb.WriteByte('\\')
			case 'n':
				sb.WriteByte('\n')
			default:
				sb.WriteRune(ch)
			}
		} else if ch == '\\' {
			escaping = true
		} else {
			sb.WriteRune(ch)
		}
	}
	if escaping {
		sb.WriteByte('\\')
	}
	return sb.String()
}

var fileNameEscapeReplacer = strings.NewReplacer("\\", `\\`, "\n", `\n`)

func escapeName(name string) (result string, changed bool) {
	if strings.IndexAny(name, "\\\n") < 0 {
		return name, false
	}
	result = fileNameEscapeReplacer.Replace(name)
	changed = true
	return
}
