package syntax

import (
	tok "hike/token"
	prs "hike/parser"
	hlv "hike/hilevel"
	hlm "hike/hilvlimpl"
)

func ParseFilesFileFilter(parser *prs.Parser) *hlm.FileTypeFilter {
	if !parser.ExpectKeyword("files") {
		return nil
	}
	parser.Next()
	return hlm.NewFileTypeFilter(false)
}

func TopFilesFileFilter(parser *prs.Parser) hlv.FileFilter {
	filter := ParseFilesFileFilter(parser)
	if filter != nil {
		return filter
	} else {
		return nil
	}
}

func ParseDirectoriesFileFilter(parser *prs.Parser) *hlm.FileTypeFilter {
	if !parser.ExpectKeyword("directories") {
		return nil
	}
	parser.Next()
	return hlm.NewFileTypeFilter(true)
}

func TopDirectoriesFileFilter(parser *prs.Parser) hlv.FileFilter {
	filter := ParseDirectoriesFileFilter(parser)
	if filter != nil {
		return filter
	} else {
		return nil
	}
}

func ParseWildcardFileFilter(parser *prs.Parser) *hlm.WildcardFileFilter {
	if !parser.ExpectKeyword("wildcard") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_STRING, "wildcard filename pattern") {
		parser.Frame("'wildcard' file filter", start)
		return nil
	}
	pattern := parser.InterpolateString()
	parser.Next()
	return hlm.NewWildcardFileFilter(pattern)
}

func TopWildcardFileFilter(parser *prs.Parser) hlv.FileFilter {
	filter := ParseWildcardFileFilter(parser)
	if filter != nil {
		return filter
	} else {
		return nil
	}
}

func ParseAnyFileFilter(parser *prs.Parser) *hlm.AnyFileFilter {
	if !parser.ExpectKeyword("any") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("'any' file filter", start)
		return nil
	}
	parser.Next()
	var children []hlv.FileFilter
	for {
		switch {
			case parser.Token.Type == tok.T_RBRACE:
				parser.Next()
				return hlm.NewAnyFileFilter(children)
			case parser.IsFileFilter():
				child := parser.FileFilter()
				if child == nil {
					parser.Frame("'any' file filter", start)
					return nil
				}
				children = append(children, child)
			default:
				parser.Die("file filter or '}'")
				parser.Frame("'any' file filter", start)
				return nil
		}
	}
}

func TopAnyFileFilter(parser *prs.Parser) hlv.FileFilter {
	filter := ParseAnyFileFilter(parser)
	if filter != nil {
		return filter
	} else {
		return nil
	}
}

func ParseAllFileFilter(parser *prs.Parser) *hlm.AllFileFilter {
	if !parser.ExpectKeyword("all") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.Expect(tok.T_LBRACE) {
		parser.Frame("'all' file filter", start)
		return nil
	}
	parser.Next()
	var children []hlv.FileFilter
	for {
		switch {
			case parser.Token.Type == tok.T_RBRACE:
				parser.Next()
				return hlm.NewAllFileFilter(children)
			case parser.IsFileFilter():
				child := parser.FileFilter()
				if child == nil {
					parser.Frame("'all' file filter", start)
					return nil
				}
				children = append(children, child)
			default:
				parser.Die("file filter or '}'")
				parser.Frame("'all' file filter", start)
				return nil
		}
	}
}

func TopAllFileFilter(parser *prs.Parser) hlv.FileFilter {
	filter := ParseAllFileFilter(parser)
	if filter != nil {
		return filter
	} else {
		return nil
	}
}

func ParseNotFileFilter(parser *prs.Parser) *hlm.NotFileFilter {
	if !parser.ExpectKeyword("not") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	child := parser.FileFilter()
	if child == nil {
		parser.Frame("'not' file filter", start)
		return nil
	}
	return hlm.NewNotFileFilter(child)
}

func TopNotFileFilter(parser *prs.Parser) hlv.FileFilter {
	filter := ParseNotFileFilter(parser)
	if filter != nil {
		return filter
	} else {
		return nil
	}
}
