package syntax

import (
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	gen "hike/generic"
	abs "hike/abstract"
	csx "hike/comsyntax"
)

func ParseCommandTransform(parser *prs.Parser) *gen.MultiCommandTransform {
	if !parser.ExpectKeyword("exec") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "transform description") {
		parser.Frame("command transform", start)
		return nil
	}
	description := parser.Token.Text
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("command transform", start)
		return nil
	}
	parser.Next()
	var words []gen.CommandWord
	for csx.IsCommandWord(parser) {
		word := csx.ParseCommandWord(parser)
		if word == nil {
			parser.Frame("command transform", start)
			return nil
		}
		words = append(words, word)
	}
	if len(words) == 0 {
		parser.Die("command word")
		parser.Frame("command transform", start)
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
	exec := gen.NewMultiCommandTransform(
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
		func(level uint) error {
			return gen.DumpCommandWords(words, level)
		},
		loud,
		suffixIsDestination,
	)
	specState := parser.SpecState()
	for parser.IsArtifactRef(true) {
		source := parser.ArtifactRef(&herr.AriseRef {
			Text: "command transform source",
			Location: &parser.Token.Location,
		}, true)
		if source == nil {
			parser.Frame("command transform", start)
			return nil
		}
		source.InjectArtifact(specState, func(artifact abs.Artifact) {
			exec.AddSource(artifact)
		})
	}
	if parser.Token.Type != tok.T_RBRACE {
		parser.Die("artifact reference or '}'")
		parser.Frame("command transform", start)
		return nil
	}
	parser.Next()
	return exec
}

func TopCommandTransform(parser *prs.Parser) abs.Transform {
	transform := ParseCommandTransform(parser)
	if transform != nil {
		return transform
	} else {
		return nil
	}
}
