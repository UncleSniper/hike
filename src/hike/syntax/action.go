package syntax

import (
	herr "hike/error"
	spc "hike/spec"
	tok "hike/token"
	prs "hike/parser"
	gen "hike/generic"
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
	ref := parser.ArtifactRef(arise, false)
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

func ParseDeleteAction(parser *prs.Parser) abs.Action {
	if !parser.ExpectKeyword("delete") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	arise := &herr.AriseRef {
		Text: "'delete' action",
		Location: start,
	}
	config := parser.SpecState().Config
	switch {
		case parser.Token.Type == tok.T_STRING:
			path := config.RealPath(parser.InterpolateString())
			parser.Next()
			action := &gen.DeletePathAction {
				Path: path,
				Base: config.TopDir,
			}
			action.Arise = arise
			return action
		case parser.IsArtifactRef(true):
			ref := parser.ArtifactRef(arise, true)
			if ref == nil {
				parser.Frame("'delete' action", start)
				return nil
			}
			action := &gen.DeleteArtifactAction {
				Base: config.TopDir,
			}
			action.Arise = arise
			ref.InjectArtifact(parser.SpecState(), func(artifact abs.Artifact) {
				action.Artifact = artifact
			})
			return action
		default:
			parser.Die("string (pathname) or artifact reference")
			parser.Frame("'delete' action", start)
			return nil
	}
}
