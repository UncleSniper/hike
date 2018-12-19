package syntax

import (
	herr "hike/error"
	spc "hike/spec"
	tok "hike/token"
	prs "hike/parser"
	gen "hike/generic"
	abs "hike/abstract"
	con "hike/concrete"
	csx "hike/comsyntax"
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
				Project: config.ProjectName,
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

func ParseCommandAction(parser *prs.Parser) abs.Action {
	if !parser.ExpectKeyword("exec") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "command description") {
		parser.Frame("command action", start)
		return nil
	}
	description := parser.InterpolateString()
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("command action", start)
		return nil
	}
	parser.Next()
	var words []gen.CommandWord
	for csx.IsCommandWord(parser) {
		word := csx.ParseCommandWord(parser)
		if word == nil {
			parser.Frame("command action", start)
			return nil
		}
		words = append(words, word)
	}
	if len(words) == 0 {
		parser.Die("command word")
		parser.Frame("command action", start)
		return nil
	}
	loud := false
	for csx.IsExecOption(parser) {
		switch parser.Token.Text {
			case "loud":
				loud = true
				parser.Next()
			case "suffixIsDestination":
				// ignored
				parser.Next()
			default:
				panic("Unrecognized exec option: " + parser.Token.Text)
		}
	}
	arise := &herr.AriseRef {
		Text: "'exec' stanza",
		Location: start,
	}
	specState := parser.SpecState()
	exec := &gen.CommandAction {
		Description: description,
		Project: specState.Config.ProjectName,
		CommandLine: func(sources []string, destinations []string) ([][]string, herr.BuildError) {
			assembled, err := gen.AssembleCommand(sources, destinations, words, arise)
			return [][]string{
				assembled,
			}, err
		},
		RequireCommandWords: func(plan *abs.Plan, arise *herr.AriseRef) herr.BuildError {
			for _, word := range words {
				rerr := word.RequireCommandWordArtifacts(plan, arise)
				if rerr != nil {
					return rerr
				}
			}
			return nil
		},
		Loud: loud,
	}
	exec.Arise = arise
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
