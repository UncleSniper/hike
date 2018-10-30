package hilvlimpl

import (
	"os"
	"io"
	"fmt"
	"path"
	"time"
	"regexp"
	"strings"
	"archive/zip"
	"path/filepath"
	herr "hike/error"
	hlv "hike/hilevel"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
)

// ---------------------------------------- BuildError ----------------------------------------

type ExtractZipError struct {
	herr.BuildErrorBase
	Archive string
	LibError error
	OperationArise *herr.AriseRef
}

func (extract *ExtractZipError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Failed to extract zip archive")
	prn.Indent(1)
	prn.Println(extract.Archive)
	prn.Indent(0)
	prn.Print("in operation ")
	prn.Arise(extract.OperationArise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Printf("because: %s", extract.LibError.Error())
	extract.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (extract *ExtractZipError) BuildErrorLocation() *loc.Location {
	return extract.OperationArise.Location
}

var _ herr.BuildError = &ExtractZipError{}

// ---------------------------------------- Step ----------------------------------------

type UnzipStep struct {
	con.StepBase
	Archives []abs.Artifact
	Destination abs.Artifact
	Valves []*UnzipValve
	Arise *herr.AriseRef
}

func (step *UnzipStep) fail(archive string, err error) herr.BuildError {
	return &ExtractZipError {
		Archive: archive,
		LibError: err,
		OperationArise: step.Arise,
	}
}

func (step *UnzipStep) Perform() herr.BuildError {
	destPaths, err := step.Destination.PathNames(nil)
	if err != nil {
		return err
	}
	if len(destPaths) != 1 {
		return &con.ConflictingDestinationsError {
			Operation: "extract zip archives to " + step.Destination.DisplayName(),
			OperationArise: step.Arise,
			PathCount: uint(len(destPaths)),
			PathsAreDestinations: true,
		}
	}
	dest := destPaths[0]
	archPaths, err := con.PathsOfArtifacts(step.Archives)
	for _, archive := range archPaths {
		finfo, nerr := os.Stat(archive)
		if nerr != nil {
			return step.fail(archive, nerr)
		}
		inf, nerr := os.Open(archive)
		if nerr != nil {
			return step.fail(archive, nerr)
		}
		zrd, nerr := zip.NewReader(inf, finfo.Size())
		if nerr != nil {
			inf.Close()
			return step.fail(archive, nerr)
		}
		for _, zfile := range zrd.File {
			fwrap := newUnzippableFile(zfile)
			for _, valve := range step.Valves {
				if !valve.Matches(fwrap) {
					continue
				}
				var newBasename, newDir, newPath string
				if valve.BasenameRegex == nil {
					newBasename = fwrap.Basename
				} else {
					newBasename = valve.BasenameRegex.ReplaceAllString(fwrap.Basename, valve.BasenameReplacement)
				}
				if strings.HasPrefix(fwrap.EnclosingDirectory, valve.RebaseFrom + "/") {
					tail := fwrap.EnclosingDirectory[len(valve.RebaseFrom):]
					newDir = filepath.Join(valve.RebaseTo, filepath.FromSlash(tail))
				} else {
					newDir = filepath.FromSlash(fwrap.EnclosingDirectory)
				}
				if filepath.IsAbs(newDir) {
					newPath = filepath.Join(newDir, newBasename)
				} else {
					newPath = filepath.Join(dest, newDir, newBasename)
				}
				if fwrap.Directory {
					nerr = os.MkdirAll(newPath, 0755)
					if nerr!= nil {
						inf.Close()
						return step.fail(archive, nerr)
					}
				} else {
					err = con.MakeEnclosingDirectories(newPath, step.Arise)
					if err != nil {
						inf.Close()
						return err
					}
					outf, nerr := os.OpenFile(newPath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
					if nerr != nil {
						inf.Close()
						return step.fail(archive, nerr)
					}
					_, nerr = io.Copy(outf, inf)
					outf.Close()
					if nerr != nil {
						inf.Close()
						return step.fail(archive, nerr)
					}
				}
				break
			}
		}
		inf.Close()
	}
	return nil
}

var _ abs.Step = &UnzipStep{}

type UnzippableFile struct {
	File *zip.File
	Path string
	Basename string
	EnclosingDirectory string
	Directory bool
}

func newUnzippableFile(file *zip.File) *UnzippableFile {
	zfname := filepath.ToSlash(file.Name)
	zfpath := path.Clean("/" + zfname)[1:]
	basename := path.Base(zfpath)
	if basename == "." || basename == "/" {
		basename = ""
	}
	dir := path.Dir(zfpath)
	if dir == "." || dir == "/" {
		dir = ""
	}
	return &UnzippableFile {
		File: file,
		Path: zfpath,
		Basename: basename,
		EnclosingDirectory: dir,
		Directory: strings.HasSuffix(zfname, "/"),
	}
}

func (file *UnzippableFile) Name() string {
	return file.Basename
}

func (file *UnzippableFile) Size() int64 {
	return int64(file.File.UncompressedSize64)
}

func (file *UnzippableFile) Mode() os.FileMode {
	if file.Directory {
		return 0755
	} else {
		return 0644
	}
}

func (file *UnzippableFile) ModTime() time.Time {
	return file.File.Modified
}

func (file *UnzippableFile) IsDir() bool {
	return file.Directory
}

func (file *UnzippableFile) Sys() interface{} {
	return nil
}

// ---------------------------------------- Transform ----------------------------------------

type UnzipValve struct {
	RebaseFrom string
	RebaseTo string
	Filters []hlv.FileFilter
	BasenameRegex *regexp.Regexp
	BasenameRegexText string
	BasenameReplacement string
}

func (valve *UnzipValve) AddFilter(filter hlv.FileFilter) {
	valve.Filters = append(valve.Filters, filter)
}

func (valve *UnzipValve) Matches(file *UnzippableFile) bool {
	switch {
		case len(file.Path) <= len(valve.RebaseFrom):
			return false
		case len(valve.RebaseFrom) > 0 && !strings.HasPrefix(file.Path, valve.RebaseFrom + "/"):
			return false
	}
	return AllFileFilters(filepath.FromSlash(file.Path), "", file, valve.Filters)
}

type UnzipTransform struct {
	con.MultiTransformBase
	Valves []*UnzipValve
	ArchiveBase string
}

func NewUnzipTransform(
	description string,
	arise *herr.AriseRef,
	archives []abs.Artifact,
	archiveBase string,
	valves []*UnzipValve,
) *UnzipTransform {
	transform := &UnzipTransform {
		Valves: valves,
		ArchiveBase: archiveBase,
	}
	transform.Sources = archives
	transform.Description = description
	transform.Arise = arise
	return transform
}

func (xform *UnzipTransform) AddArchive(archive abs.Artifact) {
	xform.Sources = append(xform.Sources, archive)
}

func (xform *UnzipTransform) AddValve(valve *UnzipValve) {
	xform.Valves = append(xform.Valves, valve)
}

func (xform *UnzipTransform) Plan(destination abs.Artifact, plan *abs.Plan) herr.BuildError {
	return con.PlanMultiTransform(
		xform,
		xform.Sources,
		destination,
		plan,
		func() herr.BuildError {
			archPaths, err := con.PathsOfArtifacts(xform.Sources)
			if err != nil {
				return err
			}
			step := &UnzipStep {
				Archives: xform.Sources,
				Destination: destination,
				Valves: xform.Valves,
				Arise: xform.Arise,
			}
			step.Description = fmt.Sprintf(
				"[%s] %s %s",
				destination.ArtifactKey().Project,
				xform.Description,
				con.GuessGroupArtifactName(archPaths, xform.ArchiveBase),
			)
			plan.AddStep(step)
			return nil
		},
	)
}

func (xform *UnzipTransform) DumpTransform(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Print("unzip ")
	con.PrintErrorString(prn, xform.Description)
	prn.Print(" {")
	for _, archive := range xform.Sources {
		prn.Println()
		prn.Indent(1)
		con.PrintErrorString(prn, archive.ArtifactKey().Unified())
	}
	for _, valve := range xform.Valves {
		prn.Println()
		prn.Indent(1)
		prn.Print("valve {")
		haveOpts := false
		if len(valve.RebaseFrom) > 0 {
			prn.Println()
			prn.Indent(2)
			prn.Print("from ")
			con.PrintErrorString(prn, valve.RebaseFrom)
			haveOpts = true
		}
		if len(valve.RebaseTo) > 0 {
			prn.Println()
			prn.Indent(2)
			prn.Print("to ")
			con.PrintErrorString(prn, valve.RebaseTo)
			haveOpts = true
		}
		if valve.BasenameRegex != nil {
			prn.Println()
			prn.Indent(2)
			prn.Print("rename ")
			con.PrintErrorString(prn, valve.BasenameRegexText)
			prn.Print(" ")
			con.PrintErrorString(prn, valve.BasenameReplacement)
			haveOpts = true
		}
		for _, filter := range valve.Filters {
			prn.Println()
			prn.Indent(2)
			prn.Inject(filter.DumpFilter, 2)
		}
		if haveOpts {
			prn.Println()
			prn.Indent(1)
		}
		prn.Print("}")
	}
	prn.Println()
	prn.Indent(0)
	prn.Print("}")
	return prn.Done()
}

var _ abs.Transform = &UnzipTransform{}
