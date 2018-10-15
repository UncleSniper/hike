package comsyntax

import (
	tok "hike/token"
	prs "hike/parser"
	gen "hike/generic"
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
