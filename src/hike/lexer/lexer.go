package lexer

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	tok "hike/token"
	loc "hike/location"
	herr "hike/error"
)

type State uint

const (
	s_NONE State = iota
	s_ERROR
	s_NAME
	s_MINUS
	s_PLUS
	s_INT
	s_STRING
	s_STRING_ESCAPE
	s_STRING_HEX
	s_STRING_UNICODE16
	s_STRING_UNICODE32
)

type Lexer struct {
	state State
	file string
	line uint
	column uint
	start uint
	buffer strings.Builder
	sink chan *tok.Token
	firstError herr.Error
	code uint32
	digits uint
}

func New(file string, sink chan *tok.Token) *Lexer {
	return &Lexer {
		file: file,
		line: 1,
		column: 1,
		sink: sink,
	}
}

type LexicalError struct {
	Unexpected rune
	Expected string
	Location *loc.Location
}

func (lerr *LexicalError) PrintError() error {
	location, err := lerr.Location.Format()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		os.Stderr,
		"Lexical error near %s at %s: Expected %s\n",
		strconv.QuoteRune(lerr.Unexpected),
		location,
		lerr.Expected,
	)
	return err
}

func (err *LexicalError) ErrorLocation() *loc.Location {
	return err.Location
}

var _ herr.Error = &LexicalError{}

func (lexer *Lexer) Location() *loc.Location {
	return &loc.Location {
		File: lexer.file,
		Line: lexer.line,
		Column: lexer.column,
	}
}

func (lexer *Lexer) die(unexpected rune, expected string) {
	lexer.firstError = &LexicalError {
		Unexpected: unexpected,
		Expected: expected,
		Location: lexer.Location(),
	}
	lexer.state = s_ERROR
}

func (lexer *Lexer) propagate(trueError error) {
	lexer.firstError = &herr.PropagatedError {
		TrueError: trueError,
		Location: lexer.Location(),
	}
	lexer.state = s_ERROR
}

func (lexer *Lexer) emitWithText(ttype tok.Type, text string) {
	lexer.sink <- &tok.Token {
		Location: loc.Location {
			File: lexer.file,
			Line: lexer.line,
			Column: lexer.start,
		},
		Type: ttype,
		Text: text,
	}
}

func (lexer *Lexer) emitFromBuffer(ttype tok.Type) {
	text := lexer.buffer.String()
	lexer.buffer.Reset()
	lexer.emitWithText(ttype, text)
}

func decodeHex(c rune) int {
	switch {
		case c >= '0' && c <= '9':
			return int(c - '0')
		case c >= 'a' && c <= 'f':
			return int(c - 'a') + 10
		case c >= 'A' && c <= 'F':
			return int(c - 'a') + 10
		default:
			return -1
	}
}

func (lexer *Lexer) doStringHex(c rune, maxDigits uint) bool {
	digit := decodeHex(c)
	if digit < 0 {
		lexer.die(c, "hexadecimal digit")
		return true
	}
	lexer.code = lexer.code * 16 + uint32(digit)
	lexer.digits++
	if lexer.digits >= maxDigits {
		_, err := lexer.buffer.WriteRune(rune(lexer.code))
		if err != nil {
			lexer.propagate(err)
			return true
		}
		lexer.digits = 0
		lexer.code = 0
	}
	return false
}

func (lexer *Lexer) PushRune(c rune) {
	var err error
  again:
	switch lexer.state {
		case s_NONE:
			lexer.start = lexer.column
			switch c {
				case ' ', '\t', '\r', '\n':
				case '{':
					lexer.emitWithText(tok.T_LBRACE, "{")
				case '}':
					lexer.emitWithText(tok.T_RBRACE, "}")
				case '-':
					lexer.state = s_MINUS
				case '+':
					lexer.state = s_PLUS
				case '"':
					lexer.state = s_STRING
				default:
					switch {
						case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
							_, err = lexer.buffer.WriteRune(c)
							if err != nil {
								lexer.propagate(err)
								return
							}
							lexer.state = s_NAME
						case c >= '0' && c <= '9':
							_, err = lexer.buffer.WriteRune(c)
							if err != nil {
								lexer.propagate(err)
								return
							}
							lexer.state = s_INT
						default:
							lexer.die(c, "whitespace or start of token")
							return
					}
			}
		case s_ERROR:
			return
		case s_NAME:
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
				_, err = lexer.buffer.WriteRune(c)
				if err != nil {
					lexer.propagate(err)
					return
				}
			} else {
				lexer.state = s_NONE
				lexer.emitFromBuffer(tok.T_NAME)
				goto again
			}
		case s_MINUS, s_PLUS:
			if c >= '0' && c <= '9' {
				_, err = lexer.buffer.WriteRune('-')
				if err != nil {
					lexer.propagate(err)
					return
				}
				_, err = lexer.buffer.WriteRune(c)
				if err != nil {
					lexer.propagate(err)
					return
				}
				lexer.state = s_INT
			} else {
				lexer.die(c, "decimal digit")
				return
			}
		case s_INT:
			if c >= '0' && c <= '9' {
				_, err = lexer.buffer.WriteRune(c)
				if err != nil {
					lexer.propagate(err)
					return
				}
			} else {
				lexer.state = s_NONE
				lexer.emitFromBuffer(tok.T_INT)
				goto again
			}
		case s_STRING:
			switch c {
				case '"':
					lexer.state = s_NONE
					lexer.emitFromBuffer(tok.T_STRING)
				case '\\':
					lexer.state = s_STRING_ESCAPE
				case '\r', '\n':
					lexer.die(c, "'\"'")
					return
				default:
					_, err = lexer.buffer.WriteRune(c)
					if err != nil {
						lexer.propagate(err)
						return
					}
			}
		case s_STRING_ESCAPE:
			switch c {
				case 'r':
					_, err = lexer.buffer.WriteRune('\r')
					lexer.state = s_STRING
				case 'n':
					_, err = lexer.buffer.WriteRune('\n')
					lexer.state = s_STRING
				case 't':
					_, err = lexer.buffer.WriteRune('\t')
					lexer.state = s_STRING
				case 'b':
					_, err = lexer.buffer.WriteRune('\b')
					lexer.state = s_STRING
				case 'a':
					_, err = lexer.buffer.WriteRune('\a')
					lexer.state = s_STRING
				case 'f':
					_, err = lexer.buffer.WriteRune('\f')
					lexer.state = s_STRING
				case 'v':
					_, err = lexer.buffer.WriteRune('\v')
					lexer.state = s_STRING
				case 'e':
					_, err = lexer.buffer.WriteRune('\033')
					lexer.state = s_STRING
				case '\\':
					_, err = lexer.buffer.WriteRune('\\')
					lexer.state = s_STRING
				case '"':
					_, err = lexer.buffer.WriteRune('"')
					lexer.state = s_STRING
				case 'x':
					lexer.state = s_STRING_HEX
				case 'u':
					lexer.state = s_STRING_UNICODE16
				case 'U':
					lexer.state = s_STRING_UNICODE32
			}
			if err != nil {
				lexer.propagate(err)
				return
			}
		case s_STRING_HEX:
			if lexer.doStringHex(c, 2) {
				return
			}
		case s_STRING_UNICODE16:
			if lexer.doStringHex(c, 4) {
				return
			}
		case s_STRING_UNICODE32:
			if lexer.doStringHex(c, 8) {
				return
			}
		default:
			panic(fmt.Sprintf("Unrecognized lexer state: %d", lexer.state))
	}
	if c == '\n' {
		lexer.line++
		lexer.column = 1
	} else {
		lexer.column++
	}
}

func (lexer *Lexer) FirstError() herr.Error {
	return lexer.firstError
}
