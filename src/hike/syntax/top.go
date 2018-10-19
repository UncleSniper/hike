package syntax

import (
	"os"
	"strconv"
	herr "hike/error"
	tok "hike/token"
	lex "hike/lexer"
	prs "hike/parser"
	abs "hike/abstract"
	con "hike/concrete"
)

func ParseGoal(parser *prs.Parser) *abs.Goal {
	if !parser.ExpectKeyword("goal") {
		return nil
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_NAME, "goal name") {
		parser.Frame("goal", start)
		return nil
	}
	goal := &abs.Goal {
		Name: parser.Token.Text,
		Arise: &herr.AriseRef {
			Text: "'goal' stanza",
			Location: start,
		},
	}
	parser.Next()
	err := parser.SpecState().RegisterGoal(goal, goal.Arise)
	if err != nil {
		parser.Fail(err)
		return nil
	}
	switch {
		case parser.IsAction():
			action := parser.Action()
			if action == nil {
				parser.Frame("goal", start)
				return nil
			}
			goal.AddAction(action)
		case parser.Token.Type == tok.T_LBRACE:
			parser.Next()
			haveLabel := false
			if parser.IsKeyword("label") {
				labelLocation := &parser.Token.Location
				parser.Next()
				if !parser.ExpectExp(tok.T_STRING, "goal label") {
					parser.Frame("'label' property", labelLocation)
					parser.Frame("goal", start)
					return nil
				}
				goal.Label = parser.InterpolateString()
				parser.Next()
				haveLabel = true
			}
		  actions:
			for {
				switch {
					case parser.IsAction():
						action := parser.Action()
						if action == nil {
							parser.Frame("goal", start)
							return nil
						}
						goal.AddAction(action)
					case parser.Token.Type == tok.T_RBRACE:
						if goal.ActionCount() == 0 {
							if haveLabel {
								parser.Die("action")
							} else {
								parser.Die("'label' or action")
							}
							parser.Frame("goal", start)
							return nil
						}
						parser.Next()
						break actions
					default:
						if haveLabel {
							if goal.ActionCount() > 0 {
								parser.Die("action or '}'")
							} else {
								parser.Die("action")
							}
						} else {
							if goal.ActionCount() > 0 {
								parser.Die("'label', action or '}'")
							} else {
								parser.Die("'label' or action")
							}
						}
						parser.Frame("goal", start)
						return nil
				}
			}
		default:
			parser.Die("action or '{'")
			parser.Frame("goal", start)
			return nil
	}
	return goal
}

func TopGoal(parser *prs.Parser) {
	ParseGoal(parser)
}

func ToplevelArtifact(parser *prs.Parser) {
	if parser.ExpectKeyword("artifact") {
		parser.Next()
		parser.Artifact()
	}
}

func ParseSetVar(parser *prs.Parser, isDef bool) {
	var initiator string
	if isDef {
		initiator = "setdef"
	} else {
		initiator = "set"
	}
	if !parser.ExpectKeyword(initiator) {
		return
	}
	start := &parser.Token.Location
	parser.Next()
	if !parser.ExpectExp(tok.T_NAME, "variable name") {
		return
	}
	name := parser.Token.Text
	parser.Next()
	switch parser.Token.Type {
		case tok.T_STRING:
			parser.SpecState().SetStringVar(name, parser.InterpolateString(), isDef)
		case tok.T_INT:
			value, err := strconv.ParseInt(parser.Token.Text, 10, 32)
			if err != nil {
				parser.Fail(&con.IllegalIntegerLiteralError {
					Specifier: parser.Token.Text,
					LibError: err,
					Location: &parser.Token.Location,
				})
				parser.Frame("variable assignment", start)
				return
			}
			parser.SpecState().SetIntVar(name, int(value), isDef)
		default:
			parser.Die("string or int")
			parser.Frame("variable assignment", start)
			return
	}
	parser.Next()
}

func TopSetVar(parser *prs.Parser) {
	ParseSetVar(parser, false)
}

func TopSetVarDef(parser *prs.Parser) {
	ParseSetVar(parser, true)
}

func includeHikefile(parser *prs.Parser, ifExists bool) herr.BuildError {
	specState := parser.SpecState()
	path := specState.Config.RealPath(parser.InterpolateString())
	pathLocation := &parser.Token.Location
	file, nerr := os.Open(path)
	if nerr != nil {
		if ifExists && os.IsNotExist(nerr) {
			parser.Next()
			return nil
		}
		return &lex.HikefileIOError {
			TrueError: nerr,
			Location: pathLocation,
		}
	}
	defer file.Close()
	tokchan := make(chan *tok.Token)
	lexer := lex.New(path, tokchan)
	oldHikefile := specState.PushHikefile(path)
	defer specState.PopHikefile(oldHikefile)
	go lexer.Slurp(file)
	oldLexer := parser.PushLexer(tokchan)
	defer parser.PopLexer(oldLexer)
	parser.Utterance()
	parser.Drain()
	parser.SupersedeError(lexer.FirstError())
	return nil
}

func ParseInclude(parser *prs.Parser) {
	if !parser.ExpectKeyword("include") {
		return
	}
	start := &parser.Token.Location
	parser.Next()
	var ifExists bool
	switch {
		case parser.Token.Type == tok.T_STRING:
		case parser.IsKeyword("ifExists"):
			ifExists = true
			parser.Next()
			if !parser.ExpectExp(tok.T_STRING, "included hikefile path") {
				parser.Frame("'include' directive", start)
				return
			}
		default:
			parser.Die("'ifExists' or string (included hikefile path)")
			parser.Frame("'include' directive", start)
			return
	}
	err := includeHikefile(parser, ifExists)
	if err != nil {
		parser.Fail(err)
		parser.Frame("'include' directive", start)
	}
}
