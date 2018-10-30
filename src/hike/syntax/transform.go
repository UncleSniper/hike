package syntax

import (
	"path"
	"regexp"
	"path/filepath"
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	gen "hike/generic"
	abs "hike/abstract"
	csx "hike/comsyntax"
	hlm "hike/hilvlimpl"
)

func ParseCommandTransform(parser *prs.Parser) *gen.MultiCommandTransform {
	if !parser.ExpectKeyword("exec") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "transform description") {
		parser.Frame("command transform", start)
		return nil
	}
	description := parser.InterpolateString()
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("command transform", start)
		return nil
	}
	parser.Next()
	var words []gen.CommandWord
	for csx.IsCommandWord(parser) {
		word := csx.ParseCommandWord(parser)
		if word == nil {
			parser.Frame("command transform", start)
			return nil
		}
		words = append(words, word)
	}
	if len(words) == 0 {
		parser.Die("command word")
		parser.Frame("command transform", start)
		return nil
	}
	loud := false
	suffixIsDestination := false
	for csx.IsExecOption(parser) {
		switch parser.Token.Text {
			case "loud":
				loud = true
				parser.Next()
			case "suffixIsDestination":
				suffixIsDestination = true
				parser.Next()
			default:
				panic("Unrecognized exec option: " + parser.Token.Text)
		}
	}
	exec := gen.NewMultiCommandTransform(
		description,
		&herr.AriseRef {
			Text: "'exec' stanza",
			Location: start,
		},
		func(sources []string, destinations []string) [][]string {
			return [][]string{
				gen.AssembleCommand(sources, destinations, words),
			}
		},
		func(level uint) error {
			return gen.DumpCommandWords(words, level)
		},
		loud,
		suffixIsDestination,
	)
	specState := parser.SpecState()
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

func TopCommandTransform(parser *prs.Parser) abs.Transform {
	transform := ParseCommandTransform(parser)
	if transform != nil {
		return transform
	} else {
		return nil
	}
}

func ParseCopyTransform(parser *prs.Parser) *gen.CopyTransform {
	if !parser.ExpectKeyword("copy") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	arise := &herr.AriseRef {
		Text: "'copy' stanza",
		Location: start,
	}
	specState := parser.SpecState()
	transform := &gen.CopyTransform {
		Sources: nil,
		UIBase: specState.Config.TopDir,
		OwningProject: specState.Config.EffectiveProjectName(),
	}
	transform.RebaseFrom = specState.Config.TopDir
	transform.Arise = arise
	switch {
		case parser.IsArtifactRef(false):
			source := parser.ArtifactRef(arise, false)
			if source == nil {
				parser.Frame("copy transform", start)
				return nil
			}
			source.InjectArtifact(specState, func(artifact abs.Artifact) {
				transform.AddSource(artifact)
			})
			return transform
		case parser.Token.Type == tok.T_LBRACE:
			parser.Next()
			initialSource := parser.ArtifactRef(arise, false)
			if initialSource == nil {
				parser.Frame("copy transform", start)
				return nil
			}
			initialSource.InjectArtifact(specState, func(artifact abs.Artifact) {
				transform.AddSource(artifact)
			})
			for parser.IsArtifactRef(false) {
				nextSource := parser.ArtifactRef(arise, false)
				if nextSource == nil {
					parser.Frame("copy transform", start)
					return nil
				}
				nextSource.InjectArtifact(specState, func(artifact abs.Artifact) {
					transform.AddSource(artifact)
				})
			}
			haveOpts := false
			for {
				switch {
					case parser.IsKeyword("rebaseFrom"):
						optloc := &parser.Token.Location
						parser.Next()
						if !parser.ExpectExp(tok.T_STRING, "pathname of base directory") {
							parser.Frame("'rebaseFrom' copy option", optloc)
							parser.Frame("copy transform", start)
							return nil
						}
						transform.RebaseFrom = specState.Config.RealPath(parser.InterpolateString())
						parser.Next()
						haveOpts = true
					case parser.IsKeyword("toDirectory"):
						transform.DestinationIsDir = true
						parser.Next()
						haveOpts = true
					case parser.Token.Type == tok.T_RBRACE:
						parser.Next()
						return transform
					default:
						if haveOpts {
							parser.Die("copy option or '}'")
						} else {
							parser.Die("artifact reference, copy option or '}'")
						}
						parser.Frame("copy transform", start)
						return nil
				}
			}
		default:
			parser.Die("artifact reference or '{'")
			parser.Frame("copy transform", start)
			return nil
	}
}

func TopCopyTransform(parser *prs.Parser) abs.Transform {
	transform := ParseCopyTransform(parser)
	if transform != nil {
		return transform
	} else {
		return nil
	}
}

func ParseZipTransform(parser *prs.Parser) *gen.ZipTransform {
	if !parser.ExpectKeyword("zip") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	var description string
	switch parser.Token.Type {
		case tok.T_STRING:
			description = parser.InterpolateString()
			parser.Next()
			if !parser.Expect(tok.T_LBRACE) {
				parser.Frame("zip transform", start)
				return nil
			}
		case tok.T_LBRACE:
		default:
			parser.Die("string (description) or '{'")
			parser.Frame("zip transform", start)
			return nil
	}
	parser.Next()
	if len(description) == 0 {
		description = "zip"
	}
	arise := &herr.AriseRef {
		Text: "'zip' stanza",
		Location: start,
	}
	transform := gen.NewZipTransform(description, arise, nil)
	specState := parser.SpecState()
	for {
		switch {
			case parser.IsKeyword("piece"):
				pstart := &parser.Token.Location
				parser.Next()
				if !parser.Expect(tok.T_LBRACE) {
					parser.Frame("zip piece", pstart)
					parser.Frame("zip transform", start)
					return nil
				}
				parser.Next()
				var rebaseFrom, rebaseTo, basenameRegexText, basenameReplacement string
				var basenameRegex *regexp.Regexp
			  pieceOpts:
				for {
					switch {
						case parser.IsKeyword("from"):
							optloc := &parser.Token.Location
							parser.Next()
							if !parser.ExpectExp(tok.T_STRING, "outer base directory") {
								parser.Frame("'from' zip piece option", optloc)
								parser.Frame("zip piece", pstart)
								parser.Frame("zip transform", start)
								return nil
							}
							rebaseFrom = specState.Config.RealPath(parser.InterpolateString())
							parser.Next()
						case parser.IsKeyword("to"):
							optloc := &parser.Token.Location
							parser.Next()
							if !parser.ExpectExp(tok.T_STRING, "inner base directory") {
								parser.Frame("'to' zip piece option", optloc)
								parser.Frame("zip piece", pstart)
								parser.Frame("zip transform", start)
								return nil
							}
							rebaseTo = path.Clean(filepath.ToSlash(parser.InterpolateString()))
							parser.Next()
						case parser.IsKeyword("rename"):
							optloc := &parser.Token.Location
							parser.Next()
							if !parser.ExpectExp(tok.T_STRING, "basename regex") {
								parser.Frame("'rename' zip piece option", optloc)
								parser.Frame("zip piece", pstart)
								parser.Frame("zip transform", start)
								return nil
							}
							var rerr error
							basenameRegexText = parser.InterpolateString()
							basenameRegex, rerr = regexp.Compile(basenameRegexText)
							if rerr != nil {
								parser.Fail(&hlm.IllegalRegexError {
									Regex: basenameRegexText,
									LibError: rerr,
									PatternArise: &herr.AriseRef {
										Text: "basename regex",
										Location: &parser.Token.Location,
									},
								})
								parser.Frame("'rename' zip piece option", optloc)
								parser.Frame("zip piece", pstart)
								parser.Frame("zip transform", start)
								return nil
							}
							parser.Next()
							if !parser.ExpectExp(tok.T_STRING, "basename replacement") {
								parser.Frame("'rename' zip piece option", optloc)
								parser.Frame("zip piece", pstart)
								parser.Frame("zip transform", start)
								return nil
							}
							basenameReplacement = parser.InterpolateString()
							parser.Next()
						default:
							break pieceOpts
					}
				}
				if len(rebaseFrom) == 0 {
					rebaseFrom = specState.Config.TopDir
				}
				piece := &gen.ZipPiece {
					RebaseFrom: rebaseFrom,
					RebaseTo: rebaseTo,
					BasenameRegex: basenameRegex,
					BasenameRegexText: basenameRegexText,
					BasenameReplacement: basenameReplacement,
				}
				haveSources := false
			  pieceArts:
				for {
					switch {
						case parser.IsArtifactRef(false):
							aref := parser.ArtifactRef(arise, false)
							if aref == nil {
								parser.Frame("zip piece", pstart)
								parser.Frame("zip transform", start)
								return nil
							}
							aref.InjectArtifact(specState, func(artifact abs.Artifact) {
								piece.AddSource(artifact)
							})
							haveSources = true
						case parser.Token.Type == tok.T_RBRACE:
							parser.Next()
							transform.AddPiece(piece)
							break pieceArts
						default:
							if haveSources {
								parser.Die("artifact reference or '}'")
							} else {
								parser.Die("zip piece option, artifact reference or '}'")
							}
							parser.Frame("zip piece", pstart)
							parser.Frame("zip transform", start)
							return nil
					}
				}
			case parser.Token.Type == tok.T_RBRACE:
				parser.Next()
				return transform
			default:
				parser.Die("zip piece or '}'")
				parser.Frame("zip transform", start)
				return nil
		}
	}
}

func TopZipTransform(parser *prs.Parser) abs.Transform {
	transform := ParseZipTransform(parser)
	if transform != nil {
		return transform
	} else {
		return nil
	}
}

func ParseUnzipTransform(parser *prs.Parser) *hlm.UnzipTransform {
	if !parser.ExpectKeyword("unzip") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	var description string
	switch parser.Token.Type {
		case tok.T_STRING:
			description = parser.InterpolateString()
			parser.Next()
			if !parser.Expect(tok.T_LBRACE) {
				parser.Frame("unzip transform", start)
				return nil
			}
		case tok.T_LBRACE:
		default:
			parser.Die("string (description) or '{'")
			parser.Frame("unzip transform", start)
			return nil
	}
	parser.Next()
	if len(description) == 0 {
		description = "unzip"
	}
	arise := &herr.AriseRef {
		Text: "'unzip' stanza",
		Location: start,
	}
	specState := parser.SpecState()
	transform := hlm.NewUnzipTransform(description, arise, nil, specState.Config.TopDir, nil)
	for {
		aref := parser.ArtifactRef(arise, false)
		if aref == nil {
			parser.Frame("unzip transform", start)
			return nil
		}
		aref.InjectArtifact(specState, func(artifact abs.Artifact) {
			transform.AddArchive(artifact)
		})
		if !parser.IsArtifactRef(false) {
			break
		}
	}
	haveValves := false
	for parser.IsKeyword("valve") {
		vstart := &parser.Token.Location
		parser.Next()
		if !parser.Expect(tok.T_LBRACE) {
			parser.Frame("unzip valve", vstart)
			parser.Frame("unzip transform", start)
			return nil
		}
		parser.Next();
		valve := &hlm.UnzipValve{}
	  valveOpts:
		for {
			switch {
				case parser.Token.Type == tok.T_RBRACE:
					parser.Next()
					transform.AddValve(valve)
					haveValves = true
					break valveOpts
				case parser.IsKeyword("from"):
					optloc := &parser.Token.Location
					parser.Next()
					if !parser.ExpectExp(tok.T_STRING, "inner base directory") {
						parser.Frame("'from' unzip valve option", optloc)
						parser.Frame("unzip valve", vstart)
						parser.Frame("unzip transform", start)
						return nil
					}
					valve.RebaseFrom = path.Clean("/" + filepath.ToSlash(parser.InterpolateString()))[1:]
					parser.Next()
				case parser.IsKeyword("to"):
					optloc := &parser.Token.Location
					parser.Next()
					if !parser.ExpectExp(tok.T_STRING, "outer base directory") {
						parser.Frame("'to' unzip valve option", optloc)
						parser.Frame("unzip valve", vstart)
						parser.Frame("unzip transform", start)
						return nil
					}
					valve.RebaseTo = specState.Config.RealPath(parser.InterpolateString())
					parser.Next()
				case parser.IsKeyword("rename"):
					optloc := &parser.Token.Location
					parser.Next()
					if !parser.ExpectExp(tok.T_STRING, "basename regex") {
						parser.Frame("'rename' unzip valve option", optloc)
						parser.Frame("unzip valve", vstart)
						parser.Frame("unzip transform", start)
						return nil
					}
					var rerr error
					valve.BasenameRegexText = parser.InterpolateString()
					valve.BasenameRegex, rerr = regexp.Compile(valve.BasenameRegexText)
					if rerr != nil {
						parser.Fail(&hlm.IllegalRegexError {
							Regex: valve.BasenameRegexText,
							LibError: rerr,
							PatternArise: &herr.AriseRef {
								Text: "basename regex",
								Location: &parser.Token.Location,
							},
						})
						parser.Frame("'rename' unzip valve option", optloc)
						parser.Frame("unzip valve", vstart)
						parser.Frame("unzip transform", start)
						return nil
					}
					parser.Next()
					if !parser.ExpectExp(tok.T_STRING, "basename replacement") {
						parser.Frame("'rename' unzip valve option", optloc)
						parser.Frame("unzip valve", vstart)
						parser.Frame("unzip transform", start)
						return nil
					}
					valve.BasenameReplacement = parser.InterpolateString()
					parser.Next()
				case parser.IsFileFilter():
					filter := parser.FileFilter()
					if filter == nil {
						parser.Frame("unzip valve", vstart)
						parser.Frame("unzip transform", start)
						return nil
					}
					valve.AddFilter(filter)
				default:
					parser.Die("unzip valve option or file filter")
					parser.Frame("unzip valve", vstart)
					parser.Frame("unzip transform", start)
					return nil
			}
		}
	}
	if parser.Token.Type != tok.T_RBRACE {
		if haveValves {
			parser.Die("unzip valve or '}'")
		} else {
			parser.Die("artifact reference, unzip valve or '}'")
		}
		parser.Frame("unzip transform", start)
		return nil
	}
	parser.Next()
	if !haveValves {
		transform.AddValve(&hlm.UnzipValve{})
	}
	return transform
}

func TopUnzipTransform(parser *prs.Parser) abs.Transform {
	transform := ParseUnzipTransform(parser)
	if transform != nil {
		return transform
	} else {
		return nil
	}
}
