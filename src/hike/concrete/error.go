package concrete

import (
	herr "hike/error"
	loc "hike/location"
	abs "hike/abstract"
)

// ---------------------------------------- CannotCanonicalizePathError ----------------------------------------

type CannotCanonicalizePathError struct {
	herr.BuildErrorBase
	Path string
	OSError error
	OperationArise *herr.AriseRef
}

func (cannot *CannotCanonicalizePathError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Failed to canonicalize path")
	prn.Indent(1)
	prn.Println(cannot.Path)
	prn.Indent(0)
	prn.Print("in operation ")
	prn.Arise(cannot.OperationArise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Printf("because: %s", cannot.OSError.Error())
	cannot.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (cannot *CannotCanonicalizePathError) BuildErrorLocation() *loc.Location {
	return cannot.OperationArise.Location
}

var _ herr.BuildError = &CannotCanonicalizePathError{}

// ---------------------------------------- CannotDeleteFileError ----------------------------------------

type CannotDeleteFileError struct {
	herr.BuildErrorBase
	Path string
	OSError error
	OperationArise *herr.AriseRef
}

func (cannot *CannotDeleteFileError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Failed to delete file")
	prn.Indent(1)
	prn.Println(cannot.Path)
	prn.Indent(0)
	prn.Print("in operation ")
	prn.Arise(cannot.OperationArise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Printf("because: %s", cannot.OSError.Error())
	cannot.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (cannot *CannotDeleteFileError) BuildErrorLocation() *loc.Location {
	return cannot.OperationArise.Location
}

var _ herr.BuildError = &CannotDeleteFileError{}

// ---------------------------------------- CannotCreateDirectoryError ----------------------------------------

type CannotCreateDirectoryError struct {
	herr.BuildErrorBase
	Path string
	OSError error
	OperationArise *herr.AriseRef
}

func (cannot *CannotCreateDirectoryError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Failed to create directory")
	prn.Indent(1)
	prn.Println(cannot.Path)
	prn.Indent(0)
	prn.Print("in operation ")
	prn.Arise(cannot.OperationArise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Printf("because: %s", cannot.OSError.Error())
	cannot.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (cannot *CannotCreateDirectoryError) BuildErrorLocation() *loc.Location {
	return cannot.OperationArise.Location
}

var _ herr.BuildError = &CannotCreateDirectoryError{}

// ---------------------------------------- UnresolvedArtifactPathError ----------------------------------------

type UnresolvedArtifactPathError struct {
	herr.BuildErrorBase
	Artifact abs.Artifact
}

func (unresolved *UnresolvedArtifactPathError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Failed to retrieve paths for artifact")
	prn.Indent(1)
	prn.Printf("%s [%s]\n", unresolved.Artifact.DisplayName(), unresolved.Artifact.ArtifactKey().Unified())
	prn.Indent(0)
	prn.Arise(unresolved.Artifact.ArtifactArise(), 0)
	prn.Println()
	prn.Indent(0)
	prn.Print("as present forward references prevent this")
	unresolved.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (unresolved *UnresolvedArtifactPathError) BuildErrorLocation() *loc.Location {
	return unresolved.Artifact.ArtifactArise().Location
}

var _ herr.BuildError = &UnresolvedArtifactPathError{}
