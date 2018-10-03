package lexer

import "os"
import "fmt"
import "strings"
import "strconv"
import tok "hike/token"
import loc "hike/location"
import herr "hike/error"

var _ = tok.T_EOF //TODO: delete
var _ = loc.Location{"", 0, 0} //TODO: delete

type State uint

const (
	s_NONE State = iota
	s_ERROR
	s_NAME
	s_INT
	s_STRING
	s_STRING_ESCAPE
	s_STRING_HEX
	s_STRING_UNICODE
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

func (lexer *Lexer) PushRune(c rune) {
	var err error
  again:
	switch lexer.state {
		case s_NONE:
			lexer.start = lexer.column
			switch c {
				case ' ', '\t', '\r', '\n':
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
				//TODO: emit token
				goto again
			}
		case s_INT:
			//TODO
		case s_STRING:
			//TODO
		case s_STRING_ESCAPE:
			//TODO
		case s_STRING_HEX:
			//TODO
		case s_STRING_UNICODE:
			//TODO
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
