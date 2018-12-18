package comsyntax

import (
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	gen "hike/generic"
	abs "hike/abstract"
)

func IsCommandWord(parser *prs.Parser) bool {
	switch parser.Token.Type {
		case tok.T_STRING, tok.T_LBRACE:
			return true
		case tok.T_NAME:
			return parser.Token.Text == "source" || parser.Token.Text == "dest" || parser.Token.Text == "aux"
		default:
			return false
	}
}

func ParseCommandWord(parser *prs.Parser) gen.CommandWord {
	switch parser.Token.Type {
		case tok.T_STRING:
			text := parser.InterpolateString()
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
					word := &gen.SourceCommandWord {}
					if parser.IsKeyword("merge") {
						word.Merge = true
						parser.Next()
					}
					return word
				case "dest":
					parser.Next()
					word := &gen.DestinationCommandWord{}
					if parser.IsKeyword("merge") {
						word.Merge = true
						parser.Next()
					}
					return word
				case "aux":
					auxloc := &parser.Token.Location
					parser.Next()
					auxart := parser.ArtifactRef(&herr.AriseRef {
						Text: "command transform auxiliary artifact",
						Location: auxloc,
					}, false)
					if auxart == nil {
						return nil
					}
					word := &gen.ArtifactCommandWord{}
					auxart.InjectArtifact(parser.SpecState(), func(artifact abs.Artifact) {
						word.Artifact = artifact
					})
					if parser.IsKeyword("merge") {
						word.Merge = true
						parser.Next()
					}
					return word
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
