package syntax

import (
	"path/filepath"
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	abs "hike/abstract"
	con "hike/concrete"
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
	switch parser.Token.Type {
		case tok.T_STRING:
			relpath := filepath.FromSlash(parser.Token.Text)
			path, nerr := filepath.Abs(relpath)
			if nerr != nil {
				parser.Fail(&con.CannotCanonicalizePathError {
					Path: relpath,
					OSError: nerr,
					OperationArise: &herr.AriseRef {
						Text: "file artifact path",
						Location: &parser.Token.Location,
					},
				})
				parser.Frame("file artifact", start)
			}
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
			relpath := parser.Token.Text
			path, nerr := filepath.Abs(relpath)
			if nerr != nil {
				parser.Fail(&con.CannotCanonicalizePathError {
					Path: relpath,
					OSError: nerr,
					OperationArise: &herr.AriseRef {
						Text: "file artifact path",
						Location: &parser.Token.Location,
					},
				})
				parser.Frame("file artifact", start)
			}
			parser.Next()
			name := ""
			haveName := false
			if parser.IsKeyword("name") {
				parser.Next()
				parser.ExpectExp(tok.T_STRING, "artifact name")
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
			case parser.IsArtifactRef():
				artifact := parser.ArtifactRef(&herr.AriseRef {
					Text: "group artifact child",
					Location: &parser.Token.Location,
				})
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
