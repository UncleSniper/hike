package syntax

import (
	herr "hike/error"
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
	if !parser.ExpectExp(tok.T_NAME, "goal name") {
		parser.Frame("attain action", start)
		return nil
	}
	specState := parser.SpecState()
	name := parser.Token.Text
	refLocation := &parser.Token.Location
	action := &con.AttainAction {
		Goal: specState.Goal(name),
	}
	action.Arise = &herr.AriseRef {
		Text: "'attain' stanza",
		Location: start,
	}
	parser.Next()
	if action.Goal == nil {
		specState.SlateResolver(func() herr.BuildError {
			action.Goal = specState.Goal(name)
			if action.Goal != nil {
				return nil
			} else {
				return &spc.NoSuchGoalError {
					Name: name,
					ReferenceLocation: refLocation,
					ReferenceArise: action.Arise,
				}
			}
		})
	}
	return action
}

func TopAttainAction(parser *prs.Parser) abs.Action {
	action := ParseAttainAction(parser)
	if action != nil {
		return action
	} else {
		return nil
	}
}

func ParseRequireAction(parser *prs.Parser) *con.RequireAction {
	if !parser.ExpectKeyword("require") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	arise := &herr.AriseRef {
		Text: "'require' stanza",
		Location: start,
	}
	ref := parser.ArtifactRef(arise)
	if ref == nil {
		parser.Frame("require action", start)
		return nil
	}
	action := &con.RequireAction {}
	action.Arise = arise
	ref.InjectArtifact(parser.SpecState(), func(artifact abs.Artifact) {
		action.Artifact = artifact
	})
	return action
}

func TopRequireAction(parser *prs.Parser) abs.Action {
	action := ParseRequireAction(parser)
	if action != nil {
		return action
	} else {
		return nil
	}
}
