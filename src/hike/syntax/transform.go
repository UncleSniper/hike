package syntax

import (
	herr "hike/error"
	//spc "hike/spec"
	tok "hike/token"
	prs "hike/parser"
	gen "hike/generic"
	abs "hike/abstract"
	//con "hike/concrete"
)

func IsCommandWord(parser *prs.Parser) bool {
	switch parser.Token.Type {
		case tok.T_STRING, tok.T_LBRACE:
			return true
		case tok.T_NAME:
			return parser.Token.Text == "source" || parser.Token.Text == "dest"
		default:
			return false
	}
}

func ParseCommandWord(parser *prs.Parser) gen.CommandWord {
	switch parser.Token.Type {
		case tok.T_STRING:
			text := parser.Token.Text
			parser.Next()
			return &gen.StaticCommandWord {
				Word: text,
			}
		case tok.T_LBRACE:
			lbrace := &parser.Token.Location
			parser.Next()
			group := &gen.BraceCommandWord{}
			for {
				switch {
					case IsCommandWord(parser):
						word := ParseCommandWord(parser)
						if word == nil {
							parser.Frame("brace comand word", lbrace)
							return nil
						}
						group.AddChild(word)
					case parser.Token.Type == tok.T_RBRACE:
						parser.Next()
						return group
					default:
						parser.Die("command word or '}'")
						parser.Frame("brace comand word", lbrace)
						return nil
				}
			}
		case tok.T_NAME:
			switch parser.Token.Text {
				case "source":
					parser.Next()
					return &gen.SourceCommandWord{}
				case "dest":
					parser.Next()
					return &gen.DestinationCommandWord{}
				default:
					parser.Die("command word")
					return nil
			}
		default:
			parser.Die("command word")
			return nil
	}
}

func IsExecOption(parser *prs.Parser) bool {
	if parser.Token.Type != tok.T_NAME {
		return false
	}
	return parser.Token.Text == "loud" || parser.Token.Text == "suffixIsDestination"
}

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
	for IsCommandWord(parser) {
		word := ParseCommandWord(parser)
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
	for IsExecOption(parser) {
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
