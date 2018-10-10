package syntax

import (
	tok "hike/token"
	prs "hike/parser"
	abs "hike/abstract"
)

func ParseGoal(parser *prs.Parser) *abs.Goal {
	if !parser.ExpectKeyword("goal") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.Expect(tok.T_NAME) {
		parser.Frame("goal", start)
		return nil
	}
	goal := &abs.Goal {
		Name: parser.Token.Text,
		Arise: &abs.AriseRef {
			Text: "'goal' stanza",
			Location: start,
		},
	}
	parser.Next()
	//TODO
	return goal
}

func TopGoal(parser *prs.Parser) {
	ParseGoal(parser)
}
