package generic

import (
	"os"
	"io"
	"fmt"
	"path/filepath"
	herr "hike/error"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
)

// ---------------------------------------- BuildError ----------------------------------------

type FileCopyIOError struct {
	herr.BuildErrorBase
	Source string
	Destination string
	OSError error
	OperationArise *herr.AriseRef
}

func (io *FileCopyIOError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Failed to copy file")
	prn.Indent(1)
	prn.Println(io.Source)
	prn.Indent(0)
	prn.Println("to")
	prn.Indent(1)
	prn.Println(io.Destination)
	prn.Indent(0)
	prn.Print("in operation ")
	prn.Arise(io.OperationArise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Printf("because: %s", io.OSError.Error())
	io.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (io *FileCopyIOError) BuildErrorLocation() *loc.Location {
	return io.OperationArise.Location
}

var _ herr.BuildError = &FileCopyIOError{}

// ---------------------------------------- Step ----------------------------------------

type CopyStep struct {
	con.StepBase
	Sources []abs.Artifact
	Destination abs.Artifact
	DestinationIsDir bool
	RebaseFrom string
	Arise *herr.AriseRef
}

func (step *CopyStep) Perform() herr.BuildError {
	destPaths, err := step.Destination.PathNames(nil)
	if err != nil {
		return err
	}
	if len(destPaths) != 1 {
		return &con.ConflictingDestinationsError {
			Operation: "copy files to: " + step.Destination.DisplayName(),
			OperationArise: step.Arise,
			PathCount: uint(len(destPaths)),
			PathsAreDestinations: true,
		}
	}
	dest := destPaths[0]
	srcPaths, err := con.PathsOfArtifacts(step.Sources)
	if err != nil {
		return err
	}
	if step.DestinationIsDir {
		var joined string
		for _, src := range srcPaths {
			rel := con.ForceToRelativeAndRebase(src, step.RebaseFrom)
			if filepath.IsAbs(rel) {
				joined = rel
			} else {
				joined = filepath.Join(dest, rel)
			}
			err = step.doCopyFile(src, joined)
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		if len(srcPaths) != 1 {
			return &con.ConflictingDestinationsError {
				Operation: "copy files to: " + step.Destination.DisplayName(),
				OperationArise: step.Arise,
				PathCount: uint(len(srcPaths)),
				PathsAreDestinations: false,
			}
		}
		return step.doCopyFile(srcPaths[0], dest)
	}
}

func (step *CopyStep) fileCopyFailed(
	source string,
	destination string,
	osError error,
) *FileCopyIOError {
	return &FileCopyIOError {
		Source: source,
		Destination: destination,
		OSError: osError,
		OperationArise: step.Arise,
	}
}

func (step *CopyStep) doCopyFile(src, dest string) herr.BuildError {
	err := con.MakeEnclosingDirectories(dest, step.Arise)
	if err != nil {
		return err
	}
	info, nerr := os.Stat(src)
	if nerr != nil {
		return step.fileCopyFailed(src, dest, nerr)
	}
	inf, nerr := os.Open(src)
	if nerr != nil {
		return step.fileCopyFailed(src, dest, nerr)
	}
	defer inf.Close()
	outf, nerr := os.OpenFile(dest, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, info.Mode() & 0777)
	if nerr != nil {
		return step.fileCopyFailed(src, dest, nerr)
	}
	_, nerr = io.Copy(outf, inf)
	outf.Close()
	if nerr != nil {
		os.Remove(dest)
		return step.fileCopyFailed(src, dest, nerr)
	}
	return nil
}

var _ abs.Step = &CopyStep{}

// ---------------------------------------- Transform ----------------------------------------

type CopyTransformBase struct {
	DestinationIsDir bool
	RebaseFrom string
	Arise *herr.AriseRef
}

type CopyTransform struct {
	CopyTransformBase
	Sources []abs.Artifact
	UIBase string
	OwningProject string
}

func (xform *CopyTransform) AddSource(source abs.Artifact) {
	xform.Sources = append(xform.Sources, source)
}

func (xform *CopyTransform) TransformDescr() string {
	return fmt.Sprintf("[%s] copy file", xform.OwningProject)
}

func (xform *CopyTransform) TransformArise() *herr.AriseRef {
	return xform.Arise
}

func (xform *CopyTransform) pathDescription(paths []string, base string) string {
	return filepath.ToSlash(con.RelPath(con.GuessGroupArtifactNameNat(paths, base), xform.UIBase))
}

func (xform *CopyTransform) Plan(destination abs.Artifact, plan *abs.Plan) herr.BuildError {
	return con.PlanMultiTransform(
		xform,
		xform.Sources,
		destination,
		plan,
		func() herr.BuildError {
			srcPaths, err := con.PathsOfArtifacts(xform.Sources)
			if err != nil {
				err.AddErrorFrame(&con.ApplyTransformFrame {
					Transform: xform,
				})
				return err
			}
			destPaths, err := destination.PathNames(nil)
			if err != nil {
				err.AddErrorFrame(&con.ApplyTransformFrame {
					Transform: xform,
				})
				return err
			}
			step := &CopyStep {
				Sources: xform.Sources,
				Destination: destination,
				DestinationIsDir: xform.DestinationIsDir,
				RebaseFrom: xform.RebaseFrom,
				Arise: xform.Arise,
			}
			step.Description = fmt.Sprintf(
				"copy %s -> %s",
				xform.pathDescription(srcPaths, xform.RebaseFrom),
				xform.pathDescription(destPaths, xform.UIBase),
			)
			plan.AddStep(step)
			return nil
		},
	)
}

func (xform *CopyTransform) DumpTransform(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Println("copy {")
	for _, source := range xform.Sources {
		prn.Indent(1)
		con.PrintErrorString(prn, source.ArtifactKey().Unified())
		prn.Println()
	}
	prn.Indent(1)
	prn.Print("rebaseFrom ")
	con.PrintErrorString(prn, xform.RebaseFrom)
	prn.Println()
	if xform.DestinationIsDir {
		prn.Indent(1)
		prn.Println("toDirectory")
	}
	prn.Indent(0)
	prn.Print("}")
	return prn.Done()
}

var _ abs.Transform = &CopyTransform{}
