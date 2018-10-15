package syntax

import (
	"fmt"
	"regexp"
	herr "hike/error"
	tok "hike/token"
	prs "hike/parser"
	hlv "hike/hilevel"
	hlm "hike/hilvlimpl"
)

func ParseStaticFile(parser *prs.Parser) *hlm.StaticFileFactory {
	if !parser.ExpectKeyword("file") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	arise := &herr.AriseRef {
		Text: "static file artifact factory",
		Location: start,
	}
	switch parser.Token.Type {
		case tok.T_STRING:
			path := parser.Token.Text
			parser.Next()
			return hlm.NewStaticFileFactory(path, "", "", "", nil, arise)
		case tok.T_LBRACE:
			parser.Next()
			if !parser.ExpectExp(tok.T_STRING, "file pathname") {
				parser.Frame("static file artifact factory", start)
				return nil
			}
			path := parser.Token.Text
			parser.Next()
			var key, name, base, optdesc string
			var optval *string
		  opts:
			for parser.Token.Type == tok.T_NAME {
				switch parser.Token.Text {
					case "key":
						optval = &key
						optdesc = "artifact key"
					case "name":
						optval = &name
						optdesc = "artifact name"
					case "base":
						optval = &base
						optdesc = "artifact base directory"
					default:
						break opts
				}
				optkey := &parser.Token.Location
				parser.Next()
				if !parser.ExpectExp(tok.T_STRING, optdesc) {
					parser.Frame(fmt.Sprintf("static file artifact factory option '%s'", optdesc), optkey)
					parser.Frame("static file artifact factory", start)
					return nil
				}
				*optval = parser.Token.Text
				parser.Next()
			}
			var generatingTransform hlv.TransformFactory
			switch {
				case parser.Token.Type == tok.T_RBRACE:
					parser.Next()
				case parser.IsTransformFactory():
					generatingTransform = parser.TransformFactory()
					if generatingTransform == nil {
						parser.Frame("static file artifact factory", start)
						return nil
					}
					parser.Expect(tok.T_RBRACE)
					parser.Next()
				default:
					parser.Die("static file artifact factory option or '}'")
					parser.Frame("static file artifact factory", start)
					return nil
			}
			return hlm.NewStaticFileFactory(path, name, key, base, generatingTransform, arise)
		default:
			parser.Die("string (file pathname) or '{'")
			parser.Frame("static file artifact factory", start)
			return nil
	}
}

func TopStaticFile(parser *prs.Parser) hlv.ArtifactFactory {
	factory := ParseStaticFile(parser)
	if factory != nil {
		return factory
	} else {
		return nil
	}
}

func manglePathRegex(parser *prs.Parser) *regexp.Regexp {
	pathRegex, rerr := regexp.Compile(parser.Token.Text)
	if rerr != nil {
		parser.Fail(&hlm.IllegalRegexError {
			Regex: parser.Token.Text,
			LibError: rerr,
			PatternArise: &herr.AriseRef {
				Text: "path regex",
				Location: &parser.Token.Location,
			},
		})
		return nil
	}
	parser.Next()
	return pathRegex
}

func ParseRegexFile(parser *prs.Parser) *hlm.RegexFileFactory {
	if !parser.ExpectKeyword("regex") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	arise := &herr.AriseRef {
		Text: "regex file artifact factory",
		Location: start,
	}
	switch parser.Token.Type {
		case tok.T_STRING:
			pathRegex := manglePathRegex(parser)
			if pathRegex == nil {
				parser.Frame("regexp file artifact factory", start)
				return nil
			}
			parser.ExpectExp(tok.T_STRING, "path replacement")
			pathReplacement := parser.Token.Text
			parser.Next()
			return hlm.NewRegexFileFactory(pathRegex, pathReplacement, "", "", "", nil, arise)
		case tok.T_LBRACE:
			parser.Next()
			if !parser.ExpectExp(tok.T_STRING, "path regex") {
				parser.Frame("regexp file artifact factory", start)
				return nil
			}
			pathRegex := manglePathRegex(parser)
			if pathRegex == nil {
				parser.Frame("regexp file artifact factory", start)
				return nil
			}
			parser.ExpectExp(tok.T_STRING, "path replacement")
			pathReplacement := parser.Token.Text
			parser.Next()
			var key, name, base, optdesc string
			var optval *string
		  opts:
			for parser.Token.Type == tok.T_NAME {
				switch parser.Token.Text {
					case "key":
						optval = &key
						optdesc = "artifact key"
					case "name":
						optval = &name
						optdesc = "artifact name"
					case "base":
						optval = &base
						optdesc = "artifact base directory"
					default:
						break opts
				}
				optkey := &parser.Token.Location
				parser.Next()
				if !parser.ExpectExp(tok.T_STRING, optdesc) {
					parser.Frame(fmt.Sprintf("regexp file artifact factory option '%s'", optdesc), optkey)
					parser.Frame("regexp file artifact factory", start)
					return nil
				}
				*optval = parser.Token.Text
				parser.Next()
			}
			var generatingTransform hlv.TransformFactory
			switch {
				case parser.Token.Type == tok.T_RBRACE:
					parser.Next()
				case parser.IsTransformFactory():
					generatingTransform = parser.TransformFactory()
					if generatingTransform == nil {
						parser.Frame("regexp file artifact factory", start)
						return nil
					}
					parser.Expect(tok.T_RBRACE)
					parser.Next()
				default:
					parser.Die("regexp file artifact factory option or '}'")
					parser.Frame("regexp file artifact factory", start)
					return nil
			}
			return hlm.NewRegexFileFactory(pathRegex, pathReplacement, name, key, base, generatingTransform, arise)
		default:
			parser.Die("string (path regex) or '{'")
			parser.Frame("regexp file artifact factory", start)
			return nil
	}
}

func TopRegexFile(parser *prs.Parser) hlv.ArtifactFactory {
	factory := ParseRegexFile(parser)
	if factory != nil {
		return factory
	} else {
		return nil
	}
}
