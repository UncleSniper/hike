package syntax

import (
	"fmt"
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
)

func ParsePipelineArtifact(parser *prs.Parser) abs.Artifact {
	if !parser.ExpectKeyword("pipeline") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("pipeline", start)
		return nil
	}
	parser.Next()
	var key, name, base, optdesc string
	var optkey *loc.Location
	var optval *string
  opts:
	for parser.Token.Type == tok.T_NAME {
		switch parser.Token.Text {
			case "key":
				optval = &key
				optdesc = "final group artifact key"
			case "name":
				optval = &name
				optdesc = "final group artifact name"
			case "base":
				optval = &base
				optdesc = "final group artifact base directory"
			default:
				break opts
		}
		optkey = &parser.Token.Location
		parser.Next()
		if !parser.ExpectExp(tok.T_STRING, optdesc) {
			parser.Frame(fmt.Sprintf("pipeline option '%s'", optdesc), optkey)
			parser.Frame("pipeline", start)
			return nil
		}
		*optval = parser.Token.Text
		parser.Next()
	}
	if !parser.IsArtifactSet() {
		parser.Die("pieline option or artifact set (initial tip)")
		parser.Frame("pipeline", start)
		return nil
	}
	tip := parser.ArtifactSet()
	if tip == nil && parser.IsError() {
		parser.Frame("pipeline", start)
		return nil
	}
	var merge bool
	specState := parser.SpecState()
	var newTip []abs.Artifact
  steps:
	for {
		switch {
			case parser.IsKeyword("merge"):
				merge = true
				parser.Next()
			case parser.IsArtifactFactory():
				merge = false
			case parser.Token.Type == tok.T_RBRACE:
				parser.Next()
				break steps
			default:
				parser.Die("'merge', artifact factory or '}'")
				parser.Frame("pipeline", start)
				return nil
		}
		step := parser.ArtifactFactory()
		if step == nil {
			parser.Frame("pipeline", start)
			return nil
		}
		if step.RequiresMerge() {
			merge = true
		}
		if merge {
			single, terr := step.NewArtifact(tip, specState)
			if terr != nil {
				parser.Fail(terr)
				parser.Frame("pipeline", start)
				return nil
			}
			tip = []abs.Artifact{single}
		} else {
			newTip = nil
			for _, artifact := range tip {
				single, terr := step.NewArtifact([]abs.Artifact{artifact}, specState)
				if terr != nil {
					parser.Fail(terr)
					parser.Frame("pipeline", start)
					return nil
				}
				newTip = append(newTip, single)
			}
			tip = newTip
		}
	}
	if len(tip) == 1 {
		return tip[0]
	}
	var allPaths []string
	for _, one := range tip {
		allPaths = one.PathNames(allPaths)
	}
	var kname, uiname string
	switch {
		case len(key) > 0:
			kname = key
		case len(base) > 0:
			kname = con.GuessGroupArtifactName(allPaths, base)
		default:
			kname = con.GuessGroupArtifactName(allPaths, specState.Config.TopDir)
	}
	switch {
		case len(name) > 0:
			uiname = name
		case len(base) > 0:
			uiname = con.GuessGroupArtifactName(allPaths, base)
		default:
			uiname = con.GuessGroupArtifactName(allPaths, specState.Config.TopDir)
	}
	gkey := abs.ArtifactKey {
		Project: specState.Config.ProjectName,
		Artifact: kname,
	}
	arise := &herr.AriseRef {
		Text: "'pipeline' stanza",
		Location: start,
	}
	group := con.NewGroup(gkey, uiname, arise)
	for _, one := range tip {
		group.AddChild(one)
	}
	err := specState.RegisterArtifact(group, arise)
	if err != nil {
		parser.Fail(err)
		parser.Frame("pipeline", start)
		return nil
	}
	return group
}
