package lexer

import (
	"io"
	"fmt"
	"bufio"
	"strings"
	"strconv"
	herr "hike/error"
	tok "hike/token"
	loc "hike/location"
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
	firstError herr.BuildError
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
	herr.BuildErrorBase
	Unexpected rune
	IsEnd bool
	Expected string
	Location *loc.Location
}

func (lerr *LexicalError) PrintBuildError(level uint) error {
	var unexpected string
	if lerr.IsEnd {
		unexpected = "end of input"
	} else {
		unexpected = strconv.QuoteRune(lerr.Unexpected)
	}
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Printf("Lexical error near %s at ", unexpected)
	prn.Location(lerr.Location)
	prn.Printf(": Expected %s\n", lerr.Expected)
	lerr.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (lerr *LexicalError) BuildErrorLocation() *loc.Location {
	return lerr.Location
}

func (err *LexicalError) ErrorLocation() *loc.Location {
	return err.Location
}

var _ herr.BuildError = &LexicalError{}

type StringBufferError struct {
	herr.BuildErrorBase
	TrueError error
	Location *loc.Location
}

func (buferr *StringBufferError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Print("Internal error in hikefile lexer at ")
	prn.Location(buferr.Location)
	prn.Printf(": Failed to write to string buffer: %s", buferr.TrueError.Error())
	buferr.InjectBacktrace(prn, level)
	return prn.Done()
}

func (buferr *StringBufferError) BuildErrorLocation() *loc.Location {
	return buferr.Location
}

var _ herr.BuildError = &StringBufferError{}

type HikefileIOError struct {
	herr.BuildErrorBase
	TrueError error
	Location *loc.Location
}

func (ioerr *HikefileIOError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Print("I/O error reading hikefile at ")
	prn.Location(ioerr.Location)
	prn.Printf(": %s", ioerr.TrueError.Error())
	ioerr.InjectBacktrace(prn, level)
	return prn.Done()
}

func (ioerr *HikefileIOError) BuildErrorLocation() *loc.Location {
	return ioerr.Location
}

var _ herr.BuildError = &HikefileIOError{}

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
		IsEnd: false,
		Expected: expected,
		Location: lexer.Location(),
	}
	lexer.state = s_ERROR
}

func (lexer *Lexer) noEnd(expected string) {
	lexer.firstError = &LexicalError {
		Unexpected: '\x00',
		IsEnd: true,
		Expected: expected,
		Location: lexer.Location(),
	}
	lexer.state = s_ERROR
}

func (lexer *Lexer) buferr(trueError error) {
	lexer.firstError = &StringBufferError {
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
			lexer.buferr(err)
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
								lexer.buferr(err)
								return
							}
							lexer.state = s_NAME
						case c >= '0' && c <= '9':
							_, err = lexer.buffer.WriteRune(c)
							if err != nil {
								lexer.buferr(err)
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
					lexer.buferr(err)
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
					lexer.buferr(err)
					return
				}
				_, err = lexer.buffer.WriteRune(c)
				if err != nil {
					lexer.buferr(err)
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
					lexer.buferr(err)
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
						lexer.buferr(err)
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
				lexer.buferr(err)
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
			panic(fmt.Sprintf("Unrecognized lexer state: %d", uint(lexer.state)))
	}
	if c == '\n' {
		lexer.line++
		lexer.column = 1
	} else {
		lexer.column++
	}
}

func (lexer *Lexer) EndUnit() herr.BuildError {
	switch lexer.state {
		case s_NONE, s_ERROR:
		case s_NAME:
			lexer.emitFromBuffer(tok.T_NAME)
		case s_MINUS, s_PLUS:
			lexer.noEnd("decimal digit")
		case s_INT:
			lexer.emitFromBuffer(tok.T_INT)
		case s_STRING:
			lexer.noEnd("'\"'")
		case s_STRING_ESCAPE:
			lexer.noEnd("escape sequence")
		case s_STRING_HEX, s_STRING_UNICODE16, s_STRING_UNICODE32:
			lexer.noEnd("hexadecimal digit")
		default:
			panic(fmt.Sprintf("Unrecognized lexer state: %d", uint(lexer.state)))
	}
	if lexer.state != s_ERROR {
		lexer.state = s_NONE
	}
	lexer.start = lexer.column
	lexer.emitWithText(tok.T_EOF, "")
	return lexer.firstError
}

func (lexer *Lexer) PushString(chunk string) {
	for _, c := range chunk {
		if lexer.state == s_ERROR {
			return
		}
		lexer.PushRune(c)
	}
}

func (lexer *Lexer) FirstError() herr.BuildError {
	return lexer.firstError
}

func (lexer *Lexer) Slurp(in io.Reader) herr.BuildError {
	br := bufio.NewReader(in)
	for lexer.state != s_ERROR {
		c, size, err := br.ReadRune()
		if size > 0 {
			lexer.PushRune(c)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			lexer.firstError = &HikefileIOError {
				TrueError: err,
				Location: lexer.Location(),
			}
			break
		}
	}
	return lexer.EndUnit()
}
