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
	if !parser.ExpectExp(tok.T_NAME, "goal name") {
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
	err := parser.SpecState().RegisterGoal(goal, goal.Arise)
	if err != nil {
		parser.Fail(err)
		return nil
	}
	switch {
		case parser.IsAction():
			action := parser.Action()
			if action == nil {
				parser.Frame("goal", start)
				return nil
			}
			goal.AddAction(action)
		case parser.Token.Type == tok.T_LBRACE:
			parser.Next()
			haveLabel := false
			if parser.IsKeyword("label") {
				labelLocation := &parser.Token.Location
				parser.Next()
				if !parser.ExpectExp(tok.T_STRING, "goal label") {
					parser.Frame("'label' property", labelLocation)
					parser.Frame("goal", start)
					return nil
				}
				goal.Label = parser.Token.Text
				parser.Next()
				haveLabel = true
			}
			for {
				switch {
					case parser.IsAction():
						action := parser.Action()
						if action == nil {
							parser.Frame("goal", start)
							return nil
						}
						goal.AddAction(action)
					case parser.Token.Type == tok.T_RBRACE:
						if goal.ActionCount() == 0 {
							if haveLabel {
								parser.Die("action")
							} else {
								parser.Die("'label' or action")
							}
							parser.Frame("goal", start)
							return nil
						}
						parser.Next()
						break
					default:
						if haveLabel {
							if goal.ActionCount() > 0 {
								parser.Die("action or '}'")
							} else {
								parser.Die("action")
							}
						} else {
							if goal.ActionCount() > 0 {
								parser.Die("'label', action or '}'")
							} else {
								parser.Die("'label' or action")
							}
						}
						parser.Frame("goal", start)
						return nil
				}
			}
		default:
			parser.Die("action or '{'")
			parser.Frame("goal", start)
			return nil
	}
	return goal
}

func TopGoal(parser *prs.Parser) {
	ParseGoal(parser)
}
