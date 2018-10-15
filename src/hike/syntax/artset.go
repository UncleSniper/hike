package syntax

import (
	tok "hike/token"
	prs "hike/parser"
	abs "hike/abstract"
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
