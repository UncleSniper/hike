package concrete

import (
	herr "hike/error"
	loc "hike/location"
	abs "hike/abstract"
)

type BuildErrorError struct {
	TrueError abs.BuildError
}

func (err *BuildErrorError) PrintError() error {
	return err.TrueError.PrintBuildError(0)
}

func (err *BuildErrorError) ErrorLocation() *loc.Location {
	return err.TrueError.BuildErrorLocation()
}

var _ herr.Error = &BuildErrorError{}
