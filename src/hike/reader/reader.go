package reader

import (
	"os"
	herr "hike/error"
	spc "hike/spec"
	tok "hike/token"
	lex "hike/lexer"
	prs "hike/parser"
	loc "hike/location"
)

func ReadFile(path string, knownStructures *prs.KnownStructures, specState *spc.State) (err herr.BuildError) {
	file, nerr := os.Open(path)
	if nerr != nil {
		err = &lex.HikefileIOError {
			TrueError: nerr,
			Location: &loc.Location {
				File: path,
				Line: 0,
				Column: 0,
			},
		}
		return
	}
	defer file.Close()
	tokchan := make(chan *tok.Token)
	lexer := lex.New(path, tokchan)
	go lexer.Slurp(file)
	parser := prs.New(tokchan, knownStructures, specState)
	parser.Utterance()
	parser.Drain()
	err = lexer.FirstError()
	if err != nil {
		return
	}
	err = parser.Error()
	return
}
