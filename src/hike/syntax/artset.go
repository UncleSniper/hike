package syntax

import (
	"os"
	"fmt"
	"path/filepath"
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	hlv "hike/hilevel"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
	hlm "hike/hilvlimpl"
)

func ParseArtifactEach(parser *prs.Parser) []abs.Artifact {
	if !parser.ExpectKeyword("each") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("artifact set literal", start)
		return nil
	}
	parser.Next()
	var set []abs.Artifact
	for {
		switch {
			case parser.Token.Type == tok.T_RBRACE:
				parser.Next()
				return set
			case parser.IsArtifactSet():
				children := parser.ArtifactSet()
				if children == nil && parser.IsError() {
					parser.Frame("artifact set literal", start)
					return nil
				}
				for _, child := range children {
					set = append(set, child)
				}
		}
	}
}

func ParseArtifactScanDir(parser *prs.Parser) []abs.Artifact {
	if !parser.ExpectKeyword("scandir") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	specState := parser.SpecState()
	switch parser.Token.Type {
		case tok.T_STRING:
			root := specState.Config.RealPath(parser.InterpolateString())
			parser.Next()
			return doArtifactScanDir(parser, start, root, "", "", "", nil)
		case tok.T_LBRACE:
			parser.Next()
			if !parser.ExpectExp(tok.T_STRING, "scan root directory") {
				parser.Frame("'scandir' artifact set", start)
				return nil
			}
			root := specState.Config.RealPath(parser.InterpolateString())
			parser.Next()
			var key, name, base, optdesc string
			var optval *string
			var isPath bool
			base = root
		  opts:
			for parser.Token.Type == tok.T_NAME {
				switch parser.Token.Text {
					case "key":
						optval = &key
						optdesc = "artifact key prefix"
						isPath = false
					case "name":
						optval = &name
						optdesc = "artifact name prefix"
						isPath = false
					case "base":
						optval = &base
						optdesc = "artifact base directory"
						isPath = true
					default:
						break opts
				}
				optkey := &parser.Token.Location
				parser.Next()
				if !parser.ExpectExp(tok.T_STRING, optdesc) {
					parser.Frame(fmt.Sprintf("'scandir' artifact set option '%s'", optdesc), optkey)
					parser.Frame("'scandir' artifact set", start)
					return nil
				}
				if isPath {
					*optval = specState.Config.RealPath(parser.InterpolateString())
				} else {
					*optval = parser.InterpolateString()
				}
				parser.Next()
			}
			var filters []hlv.FileFilter
			for parser.IsFileFilter() {
				filter := parser.FileFilter()
				if filter == nil {
					parser.Frame("'scandir' artifact set", start)
					return nil
				}
				filters = append(filters, filter)
			}
			if parser.Token.Type != tok.T_RBRACE {
				if len(filters) > 0 {
					parser.Die("file filter or '}'")
				} else {
					parser.Die("'scandir' artifact set option, file filter or '}'")
				}
				parser.Frame("'scandir' artifact set", start)
				return nil
			}
			parser.Next()
			return doArtifactScanDir(parser, start, root, key, name, base, filters)
		default:
			parser.Die("string (scan root directory) or '{'")
			parser.Frame("'scandir' artifact set", start)
			return nil
	}
}

func doArtifactScanDir(
	parser *prs.Parser,
	start *loc.Location,
	root string,
	keyPrefix string,
	namePrefix string,
	baseDir string,
	filters []hlv.FileFilter,
) []abs.Artifact {
	specState := parser.SpecState()
	config := specState.Config
	baseDir = config.RealPath(baseDir)
	arise := &herr.AriseRef {
		Text: "'scandir' artifact set",
		Location: start,
	}
	var artifacts []abs.Artifact
	outerr := filepath.Walk(root, func(fullPath string, info os.FileInfo, inerr error) error {
		if inerr != nil {
			return inerr
		}
		if !hlm.AllFileFilters(fullPath, root, info, filters) {
			return nil
		}
		name := con.GuessFileArtifactName(fullPath, baseDir)
		key := abs.ArtifactKey {
			Project: config.EffectiveProjectName(),
			Artifact: keyPrefix + name,
		}
		artifact := con.NewFile(key, namePrefix + name, arise, fullPath, nil)
		dup := specState.RegisterArtifact(artifact, arise)
		if dup != nil {
			parser.Fail(dup)
			return nil
		}
		artifacts = append(artifacts, artifact)
		return nil
	})
	if outerr != nil {
		parser.Fail(&hlm.FSWalkError {
			RootDir: root,
			OSError: outerr,
			WalkArise: arise,
		})
		return nil
	}
	return artifacts
}
