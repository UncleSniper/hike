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
	description := parser.InterpolateString()
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
	arise := &herr.AriseRef {
		Text: "'exec' stanza",
		Location: start,
	}
	exec := hlm.NewCommandTransformFactory(
		description,
		arise,
		func(sources []string, destinations []string) ([][]string, herr.BuildError) {
			assembled, err := gen.AssembleCommand(sources, destinations, words, arise)
			return [][]string{
				assembled,
			}, err
		},
		func(level uint) error {
			return gen.DumpCommandWords(words, level)
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

func ParseCopyTransformFactory(parser *prs.Parser) *hlm.CopyTransformFactory {
	if !parser.ExpectKeyword("copy") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	specState := parser.SpecState()
	arise := &herr.AriseRef {
		Text: "'copy' stanza",
		Location: start,
	}
	if parser.Token.Type != tok.T_LBRACE {
		return hlm.NewCopyTransformFactory(false, specState.Config.TopDir, arise)
	}
	parser.Next()
	rebaseFrom := specState.Config.TopDir
	toDirectory := false
	for {
		switch {
			case parser.IsKeyword("rebaseFrom"):
				optloc := &parser.Token.Location
				parser.Next()
				if !parser.ExpectExp(tok.T_STRING, "pathname of base directory") {
					parser.Frame("'rebaseFrom' copy option", optloc)
					parser.Frame("copy transform factory", start)
					return nil
				}
				rebaseFrom = specState.Config.RealPath(parser.InterpolateString())
				parser.Next()
			case parser.IsKeyword("toDirectory"):
				toDirectory = true
				parser.Next()
			case parser.Token.Type == tok.T_RBRACE:
				parser.Next()
				return hlm.NewCopyTransformFactory(toDirectory, rebaseFrom, arise)
			default:
				parser.Die("copy option or '}'")
				parser.Frame("copy transform factory", start)
				return nil
		}
	}
}

func TopCopyTransformFactory(parser *prs.Parser) hlv.TransformFactory {
	factory := ParseCopyTransformFactory(parser)
	if factory != nil {
		return factory
	} else {
		return nil
	}
}
