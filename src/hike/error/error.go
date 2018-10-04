package error

import (
	"os"
	"fmt"
	loc "hike/location"
)

type Error interface {
	PrintError() error
	ErrorLocation() *loc.Location
}

type PropagatedError struct {
	TrueError error
	Location *loc.Location
}

func (perr *PropagatedError) PrintError() error {
	location, err := perr.Location.Format()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		os.Stderr,
		"Error at %s: %s\n",
		location,
		perr.TrueError.Error(),
	)
	return err
}

func (perr *PropagatedError) ErrorLocation() *loc.Location {
	return perr.Location
}

var _ Error = &PropagatedError{}
