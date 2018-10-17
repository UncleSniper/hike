package syntax

import (
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	hlv "hike/hilevel"
	abs "hike/abstract"
	con "hike/concrete"
	hlm "hike/hilvlimpl"
)

func ParseFileArtifact(parser *prs.Parser) *con.FileArtifact {
	if !parser.ExpectKeyword("file") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "artifact key") {
		parser.Frame("file artifact", start)
		return nil
	}
	config := parser.SpecState().Config
	key := prs.SplitArtifactKey(parser.Token.Text, config)
	parser.Next()
	arise := &herr.AriseRef {
		Text: "'file' stanza",
		Location: start,
	}
	specState := parser.SpecState()
	switch parser.Token.Type {
		case tok.T_STRING:
			path := specState.Config.RealPath(parser.Token.Text)
			parser.Next()
			file := con.NewFile(*key, con.GuessFileArtifactName(path, config.TopDir), arise, path, nil)
			dup := parser.SpecState().RegisterArtifact(file, arise)
			if dup != nil {
				parser.Fail(dup)
				return nil
			}
			return file
		case tok.T_LBRACE:
			parser.Next()
			if !parser.ExpectExp(tok.T_STRING, "pathname") {
				parser.Frame("file artifact", start)
				return nil
			}
			path := specState.Config.RealPath(parser.Token.Text)
			parser.Next()
			name := ""
			haveName := false
			if parser.IsKeyword("name") {
				location := &parser.Token.Location
				parser.Next()
				if !parser.ExpectExp(tok.T_STRING, "artifact name") {
					parser.Frame("file artifact 'name' option", location)
					parser.Frame("file artifact", start)
					return nil
				}
				name = parser.Token.Text
				parser.Next()
				haveName = true
			} else {
				name = con.GuessFileArtifactName(path, config.TopDir)
			}
			var transform abs.Transform
			haveTransform := false
			if parser.IsTransform() {
				transform = parser.Transform()
				if transform == nil {
					parser.Frame("file artifact", start)
					return nil
				}
				haveTransform = true
			}
			if parser.Token.Type != tok.T_RBRACE {
				if haveName {
					if haveTransform {
						parser.Die("'}'")
					} else {
						parser.Die("transform or '}'")
					}
				} else {
					if haveTransform {
						parser.Die("'name' or '}'")
					} else {
						parser.Die("'name', transform or '}'")
					}
				}
				parser.Frame("file artifact", start)
				return nil
			}
			parser.Next()
			file := con.NewFile(*key, name, arise, path, transform)
			dup := parser.SpecState().RegisterArtifact(file, arise)
			if dup != nil {
				parser.Fail(dup)
				return nil
			}
			return file
		default:
			parser.Die("string (pathname) or '{'")
			return nil
	}
}

func TopFileArtifact(parser *prs.Parser) abs.Artifact {
	artifact := ParseFileArtifact(parser)
	if artifact != nil {
		return artifact
	} else {
		return nil
	}
}

func ParseGroupArtifact(parser *prs.Parser) *con.GroupArtifact {
	if !parser.ExpectKeyword("artifacts") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "artifact key") {
		parser.Frame("group artifact", start)
		return nil
	}
	key := prs.SplitArtifactKey(parser.Token.Text, parser.SpecState().Config)
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("group artifact", start)
		return nil
	}
	parser.Next()
	if !parser.ExpectKeyword("name") {
		parser.Frame("group artifact", start)
		return nil
	}
	nameStart := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "artifact name") {
		parser.Frame("'name' property", nameStart)
		parser.Frame("group artifact", start)
		return nil
	}
	name := parser.Token.Text
	parser.Next()
	arise := &herr.AriseRef {
		Text: "'artifacts' stanza",
		Location: start,
	}
	group := con.NewGroup(*key, name, arise)
	specState := parser.SpecState()
	dup := specState.RegisterArtifact(group, arise)
	if dup != nil {
		parser.Fail(dup)
		return nil
	}
	for {
		switch {
			case parser.Token.Type == tok.T_RBRACE:
				parser.Next()
				return group
			case parser.IsArtifactRef(false):
				artifact := parser.ArtifactRef(&herr.AriseRef {
					Text: "group artifact child",
					Location: &parser.Token.Location,
				}, false)
				if artifact == nil {
					return nil
				}
				artifact.InjectArtifact(specState, func(realArtifact abs.Artifact) {
					group.AddChild(realArtifact)
				})
			default:
				parser.Die("artifact reference or '}'")
				return nil
		}
	}
}

func TopGroupArtifact(parser *prs.Parser) abs.Artifact {
	artifact := ParseGroupArtifact(parser)
	if artifact != nil {
		return artifact
	} else {
		return nil
	}
}

func ParseTreeArtifact(parser *prs.Parser) *hlm.TreeArtifact {
	if !parser.ExpectKeyword("tree") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "artifact key") {
		parser.Frame("group artifact", start)
		return nil
	}
	specState := parser.SpecState()
	key := prs.SplitArtifactKey(parser.Token.Text, specState.Config)
	parser.Next()
	arise := &herr.AriseRef {
		Text: "'tree' stanza",
		Location: start,
	}
	switch parser.Token.Type {
		case tok.T_STRING:
			root := specState.Config.RealPath(parser.Token.Text)
			parser.Next()
			return hlm.NewTreeArtifact(
				*key,
				con.GuessFileArtifactName(root, specState.Config.TopDir),
				arise,
				root,
				nil,
				false,
			)
		case tok.T_LBRACE:
			parser.Next()
			if !parser.ExpectExp(tok.T_STRING, "root directory path") {
				parser.Frame("tree artifact", start)
				return nil
			}
			root := specState.Config.RealPath(parser.Token.Text)
			parser.Next()
			name := ""
			haveName := false
			var noCache bool
		  opts:
			for parser.Token.Type == tok.T_NAME {
				switch parser.Token.Text {
					case "name":
						location := &parser.Token.Location
						parser.Next()
						if !parser.ExpectExp(tok.T_STRING, "artifact name") {
							parser.Frame("tree artifact 'name' option", location)
							parser.Frame("tree artifact", start)
							return nil
						}
						name = parser.Token.Text
						parser.Next()
						haveName = true
					case "noCache":
						noCache = true
						parser.Next()
					default:
						break opts
				}
			}
			if !haveName {
				name = con.GuessFileArtifactName(root, specState.Config.TopDir)
			}
			var filters []hlv.FileFilter
			for parser.IsFileFilter() {
				filter := parser.FileFilter()
				if filter == nil {
					parser.Frame("tree artifact", start)
					return nil
				}
				filters = append(filters, filter)
			}
			if parser.Token.Type != tok.T_RBRACE {
				if len(filters) > 0 {
					parser.Die("file filter or '}'")
				} else {
					parser.Die("tree artifact option, file filter or '}'")
				}
				parser.Frame("tree artifact", start)
				return nil
			}
			parser.Next()
			return hlm.NewTreeArtifact(*key, name, arise, root, filters, noCache)
		default:
			parser.Die("string (root directory path) or '{'")
			parser.Frame("tree artifact", start)
			return nil
	}
}

func TopTreeArtifact(parser *prs.Parser) abs.Artifact {
	artifact := ParseTreeArtifact(parser)
	if artifact != nil {
		return artifact
	} else {
		return nil
	}
}

func ParseSplitArtifact(parser *prs.Parser) *hlm.SplitArtifact {
	if !parser.ExpectKeyword("split") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("split artifact", start)
		return nil
	}
	parser.Next()
	arise := &herr.AriseRef {
		Text: "'split' stanza",
		Location: start,
	}
	startRef := parser.ArtifactRef(arise, false)
	if startRef == nil {
		parser.Frame("split artifact", start)
		return nil
	}
	endRef := parser.ArtifactRef(arise, false)
	if endRef == nil {
		parser.Frame("split artifact", start)
		return nil
	}
	if !parser.Expect(tok.T_RBRACE) {
		parser.Frame("split artifact", start)
		return nil
	}
	parser.Next()
	split := hlm.NewSplitArtifact(nil, nil, arise)
	specState := parser.SpecState()
	startRef.InjectArtifact(specState, func(artifact abs.Artifact) {
		split.StartChild = artifact
	})
	endRef.InjectArtifact(specState, func(artifact abs.Artifact) {
		split.EndChild = artifact
	})
	return split
}

func TopSplitArtifact(parser *prs.Parser) abs.Artifact {
	artifact := ParseSplitArtifact(parser)
	if artifact != nil {
		return artifact
	} else {
		return nil
	}
}
