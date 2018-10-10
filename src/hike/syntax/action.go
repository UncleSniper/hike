package syntax

import (
	spc "hike/spec"
	tok "hike/token"
	prs "hike/parser"
	abs "hike/abstract"
	con "hike/concrete"
)

func ParseAttainAction(parser *prs.Parser) *con.AttainAction {
	if !parser.ExpectKeyword("attain") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.Expect(tok.T_NAME) {
		parser.Frame("attain action", start)
		return nil
	}
	specState := parser.SpecState()
	name := parser.Token.Text
	refLocation := &parser.Token.Location
	action := &con.AttainAction {
		Goal: specState.Goal(name),
	}
	action.Arise = &abs.AriseRef {
		Text: "'attain' stanza",
		Location: start,
	}
	parser.Next()
	if action.Goal == nil {
		specState.SlateResolver(func() abs.BuildError {
			action.Goal = specState.Goal(name)
			if action.Goal != nil {
				return nil
			} else {
				return &spc.NoSuchGoalError {
					Name: name,
					ReferenceLocation: refLocation,
					ReferenceArise: &abs.AriseRef {
						Text: "'attain' stanza",
						Location: start,
					},
				}
			}
		})
	}
	return action
}

func TopAction(parser *prs.Parser) abs.Action {
	return ParseAttainAction(parser)
}
