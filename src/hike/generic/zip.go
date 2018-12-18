package generic

import (
	"os"
	"io"
	"fmt"
	"path"
	"regexp"
	"archive/zip"
	"path/filepath"
	herr "hike/error"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
)

// ---------------------------------------- BuildError ----------------------------------------

type CreateZipError struct {
	herr.BuildErrorBase
	Destination string
	LibError error
	OperationArise *herr.AriseRef
}

func (create *CreateZipError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Failed to create zip archive")
	prn.Indent(1)
	prn.Println(create.Destination)
	prn.Indent(0)
	prn.Print("in operation ")
	prn.Arise(create.OperationArise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Printf("because: %s", create.LibError.Error())
	create.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (create *CreateZipError) BuildErrorLocation() *loc.Location {
	return create.OperationArise.Location
}

var _ herr.BuildError = &CreateZipError{}

// ---------------------------------------- Step ----------------------------------------

type ZipEmitter struct {
	archive string
	outf *os.File
	zw *zip.Writer
	knownDirs map[string]bool
	firstError herr.BuildError
}

func NewZipEmitter(archive string) (*ZipEmitter, error) {
	outf, nerr := os.OpenFile(archive, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
	if nerr != nil {
		return nil, nerr
	}
	emitter := &ZipEmitter {
		archive: archive,
		outf: outf,
		zw: zip.NewWriter(outf),
		knownDirs: make(map[string]bool),
	}
	return emitter, nil
}

func (emitter *ZipEmitter) ensureDirectory(filePath string) error {
	parent := path.Dir(filePath)
	if parent == "." {
		return nil
	}
	err := emitter.ensureDirectory(parent)
	if err != nil {
		return err
	}
	if emitter.knownDirs[parent] {
		return nil
	}
	_, err = emitter.zw.Create(parent + "/")
	if err != nil {
		return err
	}
	emitter.knownDirs[parent] = true
	return nil
}

func (emitter *ZipEmitter) EmitFile(
	name string,
	content func(io.Writer) herr.BuildError,
	fail func(error) herr.BuildError,
) {
	if emitter.firstError != nil {
		return
	}
	dest := path.Clean("/" + name)[1:]
	if len(dest) == 0 {
		return
	}
	nerr := emitter.ensureDirectory(dest)
	if nerr != nil {
		emitter.firstError = fail(nerr)
		return
	}
	into, nerr := emitter.zw.Create(dest)
	if nerr != nil {
		emitter.firstError = fail(nerr)
		return
	}
	err := content(into)
	if err != nil {
		emitter.firstError = err
		return
	}
}

func (emitter *ZipEmitter) Finish(fail func(error) herr.BuildError) herr.BuildError {
	if emitter.firstError != nil {
		emitter.zw.Close()
		emitter.outf.Close()
		os.Remove(emitter.archive)
		return emitter.firstError
	}
	nerr := emitter.zw.Close()
	if nerr != nil {
		emitter.outf.Close()
		os.Remove(emitter.archive)
		return fail(nerr)
	}
	nerr = emitter.outf.Close()
	if nerr != nil {
		os.Remove(emitter.archive)
		return fail(nerr)
	}
	return nil
}

func (emitter *ZipEmitter) Die() {
	emitter.zw.Close()
	emitter.outf.Close()
	os.Remove(emitter.archive)
}

type ZipStep struct {
	con.StepBase
	Pieces []*ZipPiece
	Destination abs.Artifact
	Arise *herr.AriseRef
}

func (step *ZipStep) fail(dest string, err error) herr.BuildError {
	return &CreateZipError {
		Destination: dest,
		LibError: err,
		OperationArise: step.Arise,
	}
}

func(step *ZipStep) Perform() herr.BuildError {
	destPaths, err := step.Destination.PathNames(nil)
	if err != nil {
		return err
	}
	if len(destPaths) != 1 {
		return &con.ConflictingDestinationsError {
			Operation: "create zip archive: " + step.Destination.DisplayName(),
			OperationArise: step.Arise,
			PathCount: uint(len(destPaths)),
			PathsAreDestinations: true,
		}
	}
	dest := destPaths[0]
	err = con.MakeEnclosingDirectories(dest, step.Arise)
	if err != nil {
		return err
	}
	emitter, nerr := NewZipEmitter(dest)
	if nerr != nil {
		return step.fail(dest, nerr)
	}
	for _, piece := range step.Pieces {
		srcPaths, err := con.PathsOfArtifacts(piece.Sources)
		if err != nil {
			emitter.Die()
			return err
		}
		for _, src := range srcPaths {
			srcTail := filepath.ToSlash(con.ForceToRelativeAndRebase(src, piece.RebaseFrom))
			destTail := filepath.ToSlash(piece.RebaseTo) + path.Clean("/" + srcTail)
			if piece.BasenameRegex != nil {
				destDirname, destBasename := path.Split(path.Clean(destTail))
				if len(destBasename) > 0 {
					newTail := piece.BasenameRegex.ReplaceAllString(destBasename, piece.BasenameReplacement)
					destTail = destDirname + newTail
				}
			}
			emitter.EmitFile(
				destTail,
				func(into io.Writer) herr.BuildError {
					inf, oserr := os.Open(src)
					if oserr != nil {
						return step.fail(dest, oserr)
					}
					_, oserr = io.Copy(into, inf)
					if oserr != nil {
						return step.fail(dest, oserr)
					}
					return nil
				},
				func(oserr error) herr.BuildError {
					return step.fail(dest, oserr)
				},
			)
		}
	}
	return emitter.Finish(func(oserr error) herr.BuildError {
		return step.fail(dest, oserr)
	})
}

var _ abs.Step = &ZipStep{}

// ---------------------------------------- Transform ----------------------------------------

type ZipPiece struct {
	Sources []abs.Artifact
	RebaseFrom string
	RebaseTo string
	BasenameRegex *regexp.Regexp
	BasenameRegexText string
	BasenameReplacement string
}

func (piece *ZipPiece) AddSource(source abs.Artifact) {
	piece.Sources = append(piece.Sources, source)
}

type ZipTransform struct {
	con.TransformBase
	Pieces []*ZipPiece
}

func NewZipTransform(description string, arise *herr.AriseRef, pieces []*ZipPiece) *ZipTransform {
	transform := &ZipTransform {
		Pieces: pieces,
	}
	transform.Description = description
	transform.Arise = arise
	return transform
}

func (xform *ZipTransform) AddPiece(piece *ZipPiece) {
	xform.Pieces = append(xform.Pieces, piece)
}

func (xform *ZipTransform) Plan(destination abs.Artifact, plan *abs.Plan) herr.BuildError {
	var sources []abs.Artifact
	for _, piece := range xform.Pieces {
		for _, source := range piece.Sources {
			sources = append(sources, source)
		}
	}
	return con.PlanMultiTransform(
		xform,
		sources,
		destination,
		plan,
		con.RequireNoMore,
		func() herr.BuildError {
			step := &ZipStep {
				Pieces: xform.Pieces,
				Destination: destination,
				Arise: xform.Arise,
			}
			step.Description = fmt.Sprintf(
				"[%s] %s %s",
				destination.ArtifactKey().Project,
				xform.Description,
				destination.DisplayName(),
			)
			plan.AddStep(step)
			return nil
		},
	)
}

func (xform *ZipTransform) DumpTransform(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Print("zip ")
	con.PrintErrorString(prn, xform.Description)
	prn.Print(" {")
	for _, piece := range xform.Pieces {
		prn.Println()
		prn.Indent(1)
		prn.Println("piece {")
		prn.Indent(2)
		prn.Print("from ")
		con.PrintErrorString(prn, piece.RebaseFrom)
		if len(piece.RebaseTo) > 0 {
			prn.Println()
			prn.Indent(2)
			prn.Print("to ")
			con.PrintErrorString(prn, piece.RebaseTo)
		}
		if piece.BasenameRegex != nil {
			prn.Println()
			prn.Indent(2)
			prn.Print("rename ")
			con.PrintErrorString(prn, piece.BasenameRegexText)
			prn.Print(" ")
			con.PrintErrorString(prn, piece.BasenameReplacement)
		}
		for _, source := range piece.Sources {
			prn.Println()
			prn.Indent(2)
			con.PrintErrorString(prn, source.ArtifactKey().Unified())
		}
		prn.Println()
		prn.Indent(1)
		prn.Print("}")
	}
	if len(xform.Pieces) > 0 {
		prn.Println()
		prn.Indent(0)
	}
	prn.Print("}")
	return prn.Done()
}

var _ abs.Transform = &ZipTransform{}
