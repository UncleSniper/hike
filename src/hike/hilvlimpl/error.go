package hilvlimpl

import (
	herr "hike/error"
	loc "hike/location"
)

type IllegalRegexError struct {
	herr.BuildErrorBase
	Regex string
	LibError error
	PatternArise *herr.AriseRef
}

func (illegal *IllegalRegexError) PrintBuildError(level uint) error {
	prn := &herr.ErrorPrinter{}
	prn.Level(level)
	prn.Println("Illegal regular expression")
	prn.Indent(1)
	prn.Println(illegal.Regex)
	prn.Indent(0)
	prn.Print("in pattern ")
	prn.Arise(illegal.PatternArise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Printf("with runtime saying: %s", illegal.LibError.Error())
	illegal.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (illegal *IllegalRegexError) BuildErrorLocation() *loc.Location {
	return illegal.PatternArise.Location
}

var _ herr.BuildError = &IllegalRegexError{}
