package hilvlimpl

import (
	"regexp"
	herr "hike/error"
	spc "hike/spec"
	gen "hike/generic"
	hlv "hike/hilevel"
	abs "hike/abstract"
	con "hike/concrete"
)

// ---------------------------------------- TransformFactory ----------------------------------------

type TransformFactoryBase struct {
	Description string
	Arise *herr.AriseRef
}

type CommandTransformFactory struct {
	TransformFactoryBase
	gen.CommandTransformBase
}

func NewCommandTransformFactory(
	description string,
	arise *herr.AriseRef,
	commandLine gen.VariableCommandLine,
	dumpCommandLine gen.CommandLineDumper,
	loud bool,
	suffixIsDestination bool,
) *CommandTransformFactory {
	factory := &CommandTransformFactory {}
	factory.Description = description
	factory.Arise = arise
	factory.CommandLine = commandLine
	factory.DumpCommandLine = dumpCommandLine
	factory.Loud = loud
	factory.SuffixIsDestination = suffixIsDestination
	return factory
}

func (factory *CommandTransformFactory) NewTransform(
	sources []abs.Artifact,
	state *spc.State,
) (abs.Transform, herr.BuildError) {
	command := gen.NewMultiCommandTransform(
		factory.Description,
		factory.Arise,
		factory.CommandLine,
		factory.DumpCommandLine,
		factory.Loud,
		factory.SuffixIsDestination,
	)
	for _, source := range sources {
		command.AddSource(source)
	}
	return command, nil
}

var _ hlv.TransformFactory = &CommandTransformFactory{}

type CopyTransformFactory struct {
	gen.CopyTransformBase
}

func NewCopyTransformFactory(
	destinationIsDir bool,
	rebaseFrom string,
	arise *herr.AriseRef,
) *CopyTransformFactory {
	factory := &CopyTransformFactory{}
	factory.DestinationIsDir = destinationIsDir
	factory.RebaseFrom = rebaseFrom
	factory.Arise = arise
	return factory
}

func (factory *CopyTransformFactory) NewTransform(
	sources []abs.Artifact,
	state *spc.State,
) (abs.Transform, herr.BuildError) {
	xform := &gen.CopyTransform {
		Sources: sources,
		UIBase: state.Config.TopDir,
		OwningProject: state.Config.EffectiveProjectName(),
	}
	xform.DestinationIsDir = factory.DestinationIsDir
	xform.RebaseFrom = factory.RebaseFrom
	xform.Arise = factory.Arise
	return xform, nil
}

var _ hlv.TransformFactory = &CopyTransformFactory{}

// ---------------------------------------- ArtifactFactory ----------------------------------------

type ArtifactFactoryBase struct {
	Arise *herr.AriseRef
}

type FileFactoryBase struct {
	BaseDir string
	GeneratingTransform hlv.TransformFactory
}

type StaticFileFactory struct {
	ArtifactFactoryBase
	FileFactoryBase
	Path string
	Name string
	Key string
}

func NewStaticFileFactory(
	path string,
	name string,
	key string,
	baseDir string,
	generatingTransform hlv.TransformFactory,
	arise *herr.AriseRef,
) *StaticFileFactory {
	factory := &StaticFileFactory {
		Path: path,
		Name: name,
		Key: key,
	}
	factory.BaseDir = baseDir
	factory.GeneratingTransform = generatingTransform
	factory.Arise = arise
	return factory
}

func (factory *StaticFileFactory) NewArtifact(
	oldArtifacts []abs.Artifact,
	state *spc.State,
) (file abs.Artifact, err herr.BuildError) {
	var kname string
	switch {
		case len(factory.Key) > 0:
			kname = factory.Key
		case len(factory.BaseDir) > 0:
			kname = con.GuessFileArtifactName(factory.Path, factory.BaseDir)
		default:
			kname = con.GuessFileArtifactName(factory.Path, state.Config.TopDir)
	}
	key := abs.ArtifactKey {
		Project: state.Config.EffectiveProjectName(),
		Artifact: kname,
	}
	var uiname string
	switch {
		case len(factory.Name) > 0:
			uiname = factory.Name
		case len(factory.BaseDir) > 0:
			uiname = con.GuessFileArtifactName(factory.Path, factory.BaseDir)
		default:
			uiname = con.GuessFileArtifactName(factory.Path, state.Config.TopDir)
	}
	var generatingTransform abs.Transform
	if factory.GeneratingTransform != nil {
		generatingTransform, err = factory.GeneratingTransform.NewTransform(oldArtifacts, state)
		if err != nil {
			return
		}
	}
	file = con.NewFile(key, uiname, factory.Arise, factory.Path, generatingTransform)
	dup := state.RegisterArtifact(file, factory.Arise)
	if dup != nil {
		err = dup
		file = nil
	}
	return
}

func (factory *StaticFileFactory) RequiresMerge() bool {
	return true
}

var _ hlv.ArtifactFactory = &StaticFileFactory{}

type RegexFileFactory struct {
	ArtifactFactoryBase
	FileFactoryBase
	PathRegex *regexp.Regexp
	PathReplacement string
	GroupName string
	GroupKey string
	RebaseFrom string
	RebaseTo string
}

func NewRegexFileFactory(
	pathRegex *regexp.Regexp,
	pathReplacement string,
	groupName string,
	groupKey string,
	rebaseFrom string,
	rebaseTo string,
	baseDir string,
	generatingTransform hlv.TransformFactory,
	arise *herr.AriseRef,
) *RegexFileFactory {
	factory := &RegexFileFactory {
		PathRegex: pathRegex,
		PathReplacement: pathReplacement,
		GroupName: groupName,
		GroupKey: groupKey,
		RebaseFrom: rebaseFrom,
		RebaseTo: rebaseTo,
	}
	factory.BaseDir = baseDir
	factory.GeneratingTransform = generatingTransform
	factory.Arise = arise
	return factory
}

func (factory *RegexFileFactory) NewArtifact(
	oldArtifacts []abs.Artifact,
	state *spc.State,
) (finalFile abs.Artifact, err herr.BuildError) {
	var files []abs.Artifact
	var allPaths, paths []string
	var kname, uiname string
	var generatingTransform abs.Transform
	shouldRebase := len(factory.RebaseFrom) > 0 || len(factory.RebaseTo) > 0
	for _, oldArtifact := range oldArtifacts {
		allPaths, err = oldArtifact.PathNames(allPaths)
		if err != nil {
			return
		}
		paths, err = oldArtifact.PathNames(nil)
		if err != nil {
			return
		}
		if factory.GeneratingTransform != nil {
			generatingTransform, err = factory.GeneratingTransform.NewTransform([]abs.Artifact{oldArtifact}, state)
			if err != nil {
				return
			}
		}
		for _, path := range paths {
			newPath := factory.PathRegex.ReplaceAllString(path, factory.PathReplacement)
			newPath = state.Config.RealPath(newPath)
			if shouldRebase {
				newPath = con.RebasePath(newPath, factory.RebaseFrom, factory.RebaseTo)
			}
			if len(factory.BaseDir) > 0 {
				kname = con.GuessFileArtifactName(newPath, factory.BaseDir)
			} else {
				kname = con.GuessFileArtifactName(newPath, state.Config.TopDir)
			}
			key := abs.ArtifactKey {
				Project: state.Config.EffectiveProjectName(),
				Artifact: kname,
			}
			file := con.NewFile(key, kname, factory.Arise, newPath, generatingTransform)
			dup := state.RegisterArtifact(file, factory.Arise)
			if dup != nil {
				err = dup
				return
			}
			files = append(files, file)
		}
	}
	if len(files) == 1 {
		finalFile = files[0]
	} else {
		switch {
			case len(factory.GroupKey) > 0:
				kname = factory.GroupKey
			case len(factory.BaseDir) > 0:
				kname = con.GuessGroupArtifactName(allPaths, factory.BaseDir)
			default:
				kname = con.GuessGroupArtifactName(allPaths, state.Config.TopDir)
		}
		key := abs.ArtifactKey {
			Project: state.Config.EffectiveProjectName(),
			Artifact: kname,
		}
		switch {
			case len(factory.GroupName) > 0:
				uiname = factory.GroupName
			case len(factory.BaseDir) > 0:
				uiname = con.GuessGroupArtifactName(allPaths, factory.BaseDir)
			default:
				uiname = con.GuessGroupArtifactName(allPaths, state.Config.TopDir)
		}
		group := con.NewGroup(key, uiname, factory.Arise)
		for _, child := range files {
			group.AddChild(child)
		}
		dup := state.RegisterArtifact(group, factory.Arise)
		if dup != nil {
			err = dup
			return
		}
		finalFile = group
	}
	return
}

func (factory *RegexFileFactory) RequiresMerge() bool {
	return false
}

var _ hlv.ArtifactFactory = &RegexFileFactory{}
