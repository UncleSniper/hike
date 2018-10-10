package parser

import (
	"fmt"
	spc "hike/spec"
	tok "hike/token"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
)

// ---------------------------------------- BuildFrame ----------------------------------------

type ParseFrame struct {
	What string
	Start *loc.Location
}

func (frame *ParseFrame) PrintErrorFrame(level uint) error {
	prn := &abs.ErrorPrinter{}
	prn.Printf("parsing %s starting at ", frame.What)
	prn.Location(frame.Start)
	return prn.Done()
}

var _ abs.BuildFrame = &ParseFrame{}

// ---------------------------------------- BuildError ----------------------------------------

type SyntaxError struct {
	con.BuildErrorBase
	Near *tok.Token
	Expected string
}

func (syntax *SyntaxError) PrintBuildError(level uint) error {
	prn := &abs.ErrorPrinter{}
	prn.Level(level)
	prn.Print("Syntax error at ")
	prn.Location(&syntax.Near.Location)
	prn.Println()
	prn.Indent(1)
	prn.Inject(func(innerLevel uint) error {
		str, err := syntax.Near.Reconstruct()
		if err != nil {
			return err
		}
		prn.Print("found: ", str)
		return nil
	}, 0)
	prn.Println()
	prn.Indent(1)
	prn.Print("expected: ", syntax.Expected)
	syntax.InjectBacktrace(prn, 0)
	return prn.Done()
}

var _ abs.BuildError = &SyntaxError{}

// ---------------------------------------- Parser ----------------------------------------

type Parser struct {
	lexer chan *tok.Token
	Token *tok.Token
	firstError abs.BuildError
	knownStructures *KnownStructures
	specState *spc.State
}

type TopParser func(parser *Parser)
type ActionParser func(parser *Parser) abs.Action

type KnownStructures struct {
	top map[string]TopParser
	action map[string]ActionParser
}

func NewKnownStructures() *KnownStructures {
	return &KnownStructures {
		top: make(map[string]TopParser),
		action: make(map[string]ActionParser),
	}
}

func (known *KnownStructures) RegisterTopParser(initiator string, parser TopParser) {
	known.top[initiator] = parser
}

func (known *KnownStructures) TopParser(initiator string) TopParser {
	return known.top[initiator]
}

func (known *KnownStructures) RegisterActionParser(initiator string, parser ActionParser) {
	known.action[initiator] = parser
}

func (known *KnownStructures) ActionParser(initiator string) ActionParser {
	return known.action[initiator]
}

func New(
	lexer chan *tok.Token,
	knownStructures *KnownStructures,
	specState *spc.State,
) *Parser {
	return &Parser {
		lexer: lexer,
		Token: <-lexer,
		knownStructures: knownStructures,
		specState: specState,
	}
}

func (parser *Parser) Next() {
	if parser.Token.Type != tok.T_EOF {
		parser.Token = <-parser.lexer
	}
}

func (parser *Parser) Die(expected string) {
	if parser.firstError != nil {
		parser.firstError = &SyntaxError {
			Near: parser.Token,
			Expected: expected,
		}
	}
}

func (parser *Parser) Expect(ttype tok.Type) bool {
	if parser.Token.Type == ttype {
		return true
	}
	parser.Die(tok.NameType(ttype))
	return false
}

func (parser *Parser) ExpectKeyword(name string) bool {
	if parser.Token.Type == tok.T_NAME && parser.Token.Text == name {
		return true
	}
	parser.Die(fmt.Sprintf("'%s'", name))
	return false
}

func (parser *Parser) Fail(fault abs.BuildError) {
	if parser.firstError == nil {
		parser.firstError = fault
	}
}

func (parser *Parser) Frame(what string, start *loc.Location) {
	if start == nil {
		start = &parser.Token.Location
	}
	if parser.firstError != nil {
		parser.firstError.AddErrorFrame(&ParseFrame {
			What: what,
			Start: start,
		})
	}
}

func (parser *Parser) Top() {
	if !parser.Expect(tok.T_NAME) {
		return
	}
	cb := parser.knownStructures.TopParser(parser.Token.Text)
	if cb == nil {
		parser.Die("top-level definition")
	} else {
		cb(parser)
	}
}

func (parser *Parser) Action() abs.Action {
	if !parser.Expect(tok.T_NAME) {
		return nil
	}
	cb := parser.knownStructures.ActionParser(parser.Token.Text)
	if cb == nil {
		parser.Die("action")
		return nil
	} else {
		return cb(parser)
	}
}
