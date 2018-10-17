package hilvlimpl

import (
	herr "hike/error"
	loc "hike/location"
)

// ---------------------------------------- IllegalRegexError ----------------------------------------

type IllegalRegexError struct {
	herr.BuildErrorBase
	Regex string
	LibError error
	PatternArise *herr.AriseRef
}

func (illegal *IllegalRegexError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
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

// ---------------------------------------- FSWalkError ----------------------------------------

type FSWalkError struct {
	herr.BuildErrorBase
	RootDir string
	OSError error
	WalkArise *herr.AriseRef
}

func (walk *FSWalkError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Failed to walk filesystem tree")
	prn.Indent(1)
	prn.Println(walk.RootDir)
	prn.Indent(0)
	prn.Print("in operation ")
	prn.Arise(walk.WalkArise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Printf("because: %s", walk.OSError.Error())
	walk.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (walk *FSWalkError) BuildErrorLocation() *loc.Location {
	return walk.WalkArise.Location
}

var _ herr.BuildError = &FSWalkError{}
