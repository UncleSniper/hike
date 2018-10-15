package syntax

import (
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	gen "hike/generic"
	hlv "hike/hilevel"
	hlm "hike/hilvlimpl"
	csx "hike/comsyntax"
)

func ParseCommandTransformFactory(parser *prs.Parser) *hlm.CommandTransformFactory {
	if !parser.ExpectKeyword("exec") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "transform description") {
		parser.Frame("command transform factory", start)
		return nil
	}
	description := parser.Token.Text
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("command transform factory", start)
		return nil
	}
	parser.Next()
	var words []gen.CommandWord
	for csx.IsCommandWord(parser) {
		word := csx.ParseCommandWord(parser)
		if word == nil {
			parser.Frame("command transform factory", start)
			return nil
		}
		words = append(words, word)
	}
	if len(words) == 0 {
		parser.Die("command word")
		parser.Frame("command transform factory", start)
		return nil
	}
	loud := false
	suffixIsDestination := false
	for csx.IsExecOption(parser) {
		switch parser.Token.Text {
			case "loud":
				loud = true
				parser.Next()
			case "suffixIsDestination":
				suffixIsDestination = true
				parser.Next()
			default:
				panic("Unrecognized exec option: " + parser.Token.Text)
		}
	}
	exec := hlm.NewCommandTransformFactory(
		description,
		&herr.AriseRef {
			Text: "'exec' stanza",
			Location: start,
		},
		func(sources []string, destinations []string) [][]string {
			return [][]string{
				gen.AssembleCommand(sources, destinations, words),
			}
		},
		loud,
		suffixIsDestination,
	)
	if parser.Token.Type != tok.T_RBRACE {
		parser.Die("command option or '}'")
		parser.Frame("command transform factory", start)
		return nil
	}
	parser.Next()
	return exec
}

func TopCommandTransformFactory(parser *prs.Parser) hlv.TransformFactory {
	factory := ParseCommandTransformFactory(parser)
	if factory != nil {
		return factory
	} else {
		return nil
	}
}
