package reader

import (
	"os"
	herr "hike/error"
	/*
	tok "hike/token"
	lex "hike/lexer"
	prs "hike/parser"
	con "hike/concrete"
	*/
)

func ReadFile(path string) (err herr.BuildError) {
	file, nerr := os.Open(path)
	if nerr != nil {
		//err = 
		return
	}
	defer file.Close()
	//TODO
	return
}
