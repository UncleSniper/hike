package token

import (
	"fmt"
	"strings"
	"strconv"
	loc "hike/location"
)

type Type uint

const (
	T_NAME Type = iota
	T_STRING
	T_INT
	T_LBRACE
	T_RBRACE
	T_EOF
)

type Token struct {
	Location loc.Location
	Type Type
	Text string
}

func EscapeRuneTo(c rune, quote bool, delimiter rune, sink *strings.Builder) (err error) {
	if quote {
		_, err = sink.WriteRune('\'')
		if err != nil {
			return
		}
	}
	switch c {
		case '\r':
			_, err = sink.WriteString("\\r")
		case '\n':
			_, err = sink.WriteString("\\n")
		case '\t':
			_, err = sink.WriteString("\\t")
		case '\b':
			_, err = sink.WriteString("\\b")
		case '\a':
			_, err = sink.WriteString("\\a")
		case '\f':
			_, err = sink.WriteString("\\f")
		case '\v':
			_, err = sink.WriteString("\\v")
		case '\033':
			_, err = sink.WriteString("\\e")
		case '\\':
			_, err = sink.WriteString("\\\\")
		default:
			switch {
				case c == delimiter:
					_, err = sink.WriteRune('\\')
					if err != nil {
						return
					}
					_, err = sink.WriteRune(c)
				case strconv.IsPrint(c):
					_, err = sink.WriteRune(c)
				case c <= 255:
					_, err = sink.WriteString(fmt.Sprintf("\\x%02X", uint(c)))
				case c <= 0xFFFF:
					_, err = sink.WriteString(fmt.Sprintf("\\u%04X", uint(c)))
				default:
					_, err = sink.WriteString(fmt.Sprintf("\\U%08X", uint32(c)))
			}
	}
	if err != nil {
		return
	}
	if quote {
		_, err = sink.WriteRune('\'')
	}
	return
}

func EscapeRune(c rune, quote bool, delimiter rune) (escaped string, err error) {
	var sink strings.Builder
	err = EscapeRuneTo(c, quote, delimiter, &sink)
	escaped = sink.String()
	return
}

func EscapeStringTo(text string, quote bool, sink *strings.Builder) (err error) {
	if quote {
		_, err = sink.WriteRune('"')
		if err != nil {
			return
		}
	}
	for _, c := range text {
		err = EscapeRuneTo(c, false, '"', sink)
		if err != nil {
			return
		}
	}
	if quote {
		_, err = sink.WriteRune('"')
	}
	return
}

func EscapeString(text string, quote bool) (escaped string, err error) {
	var sink strings.Builder
	err = EscapeStringTo(text, quote, &sink)
	escaped = sink.String()
	return
}

func (token *Token) ReconstructTo(sink *strings.Builder) (err error) {
	switch token.Type {
		case T_STRING:
			err = EscapeStringTo(token.Text, true, sink)
		case T_EOF:
			_, err = sink.WriteString("<end of input>")
		default:
			_, err = sink.WriteRune('\'')
			if err != nil {
				return
			}
			_, err = sink.WriteString(token.Text)
			if err != nil {
				return
			}
			_, err = sink.WriteRune('\'')
	}
	return
}

func (token *Token) Reconstruct() (text string, err error) {
	var sink strings.Builder
	err = token.ReconstructTo(&sink)
	text = sink.String()
	return
}

func NameType(ttype Type) string {
	switch ttype {
		case T_NAME:
			return "name"
		case T_STRING:
			return "string"
		case T_INT:
			return "int"
		case T_LBRACE:
			return "'{'"
		case T_RBRACE:
			return "'}'"
		case T_EOF:
			return "end of input"
		default:
			panic(fmt.Sprintf("Unrecognized token type: %d", uint(ttype)))
	}
}
