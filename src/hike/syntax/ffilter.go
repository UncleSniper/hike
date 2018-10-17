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
	pattern := parser.Token.Text
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
