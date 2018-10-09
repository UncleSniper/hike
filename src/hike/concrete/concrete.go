package concrete

import (
	"os"
	"fmt"
	"time"
	abs "hike/abstract"
)

// ---------------------------------------- BuildFrame ----------------------------------------

func PrintArtifactErrorFrameBase(level uint, action string, artifact abs.Artifact) error {
	prn := &abs.ErrorPrinter{}
	prn.Level(level)
	prn.Printf("%s artifact\n", action)
	prn.Indent(1)
	prn.Printf(
		"%s [%s]\n",
		artifact.DisplayName(),
		artifact.ArtifactKey().Unified(),
	)
	prn.Indent(0)
	prn.Arise(artifact.ArtifactArise(), 0)
	return prn.Done()
}

type RequireArtifactFrame struct {
	Artifact abs.Artifact
}

func (frame *RequireArtifactFrame) PrintErrorFrame(level uint) error {
	return PrintArtifactErrorFrameBase(level, "requiring", frame.Artifact)
}

var _ abs.BuildFrame = &RequireArtifactFrame{}

type FlattenArtifactFrame struct {
	Artifact abs.Artifact
}

func (frame *FlattenArtifactFrame) PrintErrorFrame(level uint) error {
	return PrintArtifactErrorFrameBase(level, "flattening", frame.Artifact)
}

var _ abs.BuildFrame = &FlattenArtifactFrame{}

type ApplyTransformFrame struct {
	Transform abs.Transform
}

func (frame *ApplyTransformFrame) PrintErrorFrame(level uint) (err error) {
	_, err = fmt.Fprintln(os.Stderr, "applying transform")
	if err != nil {
		return
	}
	err = abs.IndentError(level + 1)
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(os.Stderr, frame.Transform.TransformDescr())
	if err != nil {
		return
	}
	err = abs.IndentError(level)
	if err != nil {
		return
	}
	err = frame.Transform.TransformArise().PrintArise(level)
	return
}

var _ abs.BuildFrame = &ApplyTransformFrame{}

type AttainGoalFrame struct {
	Goal *abs.Goal
}

func (frame *AttainGoalFrame) PrintErrorFrame(level uint) (err error) {
	_, err = fmt.Fprintf(os.Stderr, "attaining goal '%s' ", frame.Goal.Name)
	if err != nil {
		return
	}
	err = frame.Goal.Arise.PrintArise(level)
	return
}

var _ abs.BuildFrame = &AttainGoalFrame{}

type PerformActionFrame struct {
	Action abs.Action
}

func (frame *PerformActionFrame) PrintErrorFrame(level uint) (err error) {
	_, err = fmt.Fprintf(os.Stderr, "performing action '%s' ", frame.Action.SimpleDescr())
	if err != nil {
		return
	}
	err = frame.Action.ActionArise().PrintArise(level)
	return
}

// ---------------------------------------- BuildError ----------------------------------------

type BuildErrorBase struct {
	frames []abs.BuildFrame
}

func (err *BuildErrorBase) AddErrorFrame(frame abs.BuildFrame) {
	err.frames = append(err.frames, frame)
}

func (base *BuildErrorBase) PrintBacktrace(level uint) (err error) {
	count := len(base.frames)
	for i := 0; i < count; i++ {
		_, err = fmt.Fprintln(os.Stderr)
		if err != nil {
			return
		}
		err = abs.IndentError(level)
		if err != nil {
			return
		}
		err = base.frames[i].PrintErrorFrame(level + 1)
		if err != nil {
			return
		}
	}
	return
}

type NoGeneratorError struct {
	BuildErrorBase
	Artifact abs.Artifact
}

func (nogen *NoGeneratorError) PrintBuildError(level uint) (err error) {
	_, err = fmt.Fprintln(os.Stderr, "Don't know how to obtain artifact")
	if err != nil {
		return
	}
	err = abs.IndentError(level + 1)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(
		os.Stderr,
		"%s [%s]\n",
		nogen.Artifact.DisplayName(),
		nogen.Artifact.ArtifactKey().Unified(),
	)
	if err != nil {
		return
	}
	err = abs.IndentError(level)
	if err != nil {
		return
	}
	err = nogen.Artifact.ArtifactArise().PrintArise(level)
	if err != nil {
		return
	}
	err = nogen.PrintBacktrace(level)
	return
}

var _ abs.BuildError = &NoGeneratorError{}

type CannotStatError struct {
	BuildErrorBase
	Path string
	OSError error
}

func (cannot *CannotStatError) PrintBuildError(level uint) (err error) {
	_, err = fmt.Fprintln(os.Stderr, "Failed to stat file")
	if err != nil {
		return
	}
	err = abs.IndentError(level + 1)
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(os.Stderr, cannot.Path)
	if err != nil {
		return
	}
	err = abs.IndentError(level)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(os.Stderr, "because: %s", cannot.OSError.Error())
	if err != nil {
		return
	}
	err = cannot.PrintBacktrace(level)
	return
}

var _ abs.BuildError = &CannotStatError{}

// ---------------------------------------- Artifact ----------------------------------------

type ArtifactBase struct {
	Key abs.ArtifactKey
	ID abs.ArtifactID
	Name string
	Arise *abs.AriseRef
}

func (artifact *ArtifactBase) ArtifactKey() *abs.ArtifactKey {
	return &artifact.Key
}

func (artifact *ArtifactBase) ArtifactID() abs.ArtifactID {
	return artifact.ID
}

func (artifact *ArtifactBase) ArtifactArise() *abs.AriseRef {
	return artifact.Arise
}

type FileArtifact struct {
	ArtifactBase
	Path string
	GeneratingTransform abs.Transform
}

func (artifact *FileArtifact) DisplayName() string {
	if len(artifact.Name) > 0 {
		return artifact.Name
	} else {
		return artifact.Path
	}
}

func (artifact *FileArtifact) PathNames(sink []string) []string {
	return append(sink, artifact.Path)
}

func (artifact *FileArtifact) ModTime() (stamp time.Time, err abs.BuildError, missing bool) {
	info, oserr := os.Stat(artifact.Path)
	if oserr == nil {
		stamp = info.ModTime()
	} else {
		missing = os.IsNotExist(oserr)
		if !missing {
			err = &CannotStatError {
				Path: artifact.Path,
				OSError: oserr,
			}
		}
	}
	return
}

func (artifact *FileArtifact) EarliestModTime() (time.Time, abs.BuildError, bool) {
	return artifact.ModTime()
}

func (artifact *FileArtifact) LatestModTime() (time.Time, abs.BuildError, bool) {
	return artifact.ModTime()
}

func (artifact *FileArtifact) Flatten() abs.BuildError {
	return nil
}

func FileExists(path string) (exists bool, err abs.BuildError) {
	_, oserr := os.Lstat(path)
	switch {
		case oserr == nil:
			exists = true
		case !os.IsNotExist(oserr):
			err = &CannotStatError {
				Path: path,
				OSError: oserr,
			}
	}
	return
}

func (artifact *FileArtifact) Require(plan *abs.Plan) (err abs.BuildError) {
	if plan.AlreadyUpToDate(artifact) {
		return
	}
	if artifact.GeneratingTransform != nil {
		err = artifact.GeneratingTransform.Plan(artifact, plan)
		if err != nil {
			err.AddErrorFrame(&RequireArtifactFrame {
				Artifact: artifact,
			})
		}
	} else {
		exists, err := FileExists(artifact.Path)
		switch {
			case err != nil:
				err.AddErrorFrame(&RequireArtifactFrame {
					Artifact: artifact,
				})
			case !exists:
				err = &NoGeneratorError {
					Artifact: artifact,
				}
		}
	}
	if err != nil {
		plan.BroughtUpToDate(artifact)
	}
	return
}

var _ abs.Artifact = &FileArtifact{}

func NewFile(
	key abs.ArtifactKey,
	name string,
	arise *abs.AriseRef,
	path string,
	generatingTransform abs.Transform,
) *FileArtifact {
	artifact := &FileArtifact {
		Path: path,
		GeneratingTransform: generatingTransform,
	}
	artifact.Key = key
	artifact.ID = abs.NextArtifactID()
	artifact.Name = name
	artifact.Arise = arise
	return artifact
}

type GroupArtifact struct {
	ArtifactBase
	children []abs.Artifact
}

func (group *GroupArtifact) AddChild(child abs.Artifact) {
	group.children = append(group.children, child)
}

func (artifact *GroupArtifact) DisplayName() string {
	return artifact.Name
}

func (artifact *GroupArtifact) PathNames(sink []string) []string {
	for _, child := range artifact.children {
		sink = child.PathNames(sink)
	}
	return sink
}

func (artifact *GroupArtifact) EarliestModTime() (result time.Time, err abs.BuildError, missing bool) {
	result = time.Now()
	var (
		have bool
		chmod time.Time
		chmiss bool
	)
	for _, child := range artifact.children {
		chmod, err, chmiss = child.EarliestModTime()
		missing = missing || chmiss
		if err != nil {
			return
		}
		switch {
			case !have:
				result = chmod
				have = true
			case result.After(chmod):
				result = chmod
		}
	}
	return
}

func (artifact *GroupArtifact) LatestModTime() (result time.Time, err abs.BuildError, missing bool) {
	result = time.Now()
	var (
		have bool
		chmod time.Time
		chmiss bool
	)
	for _, child := range artifact.children {
		chmod, err, chmiss = child.LatestModTime()
		missing = missing || chmiss
		if err != nil {
			return
		}
		switch {
			case !have:
				result = chmod
				have = true
			case chmod.After(result):
				result = chmod
		}
	}
	return
}

func (artifact *GroupArtifact) Flatten() (err abs.BuildError) {
	for _, child := range artifact.children {
		err = child.Flatten()
		if err != nil {
			err.AddErrorFrame(&FlattenArtifactFrame {
				Artifact: artifact,
			})
			return
		}
	}
	return
}

func (artifact *GroupArtifact) Require(plan *abs.Plan) (err abs.BuildError) {
	if plan.AlreadyUpToDate(artifact) {
		return
	}
	for _, child := range artifact.children {
		err = child.Require(plan)
		if err != nil {
			err.AddErrorFrame(&RequireArtifactFrame {
				Artifact: artifact,
			})
			return
		}
	}
	plan.BroughtUpToDate(artifact)
	return
}

var _ abs.Artifact = &GroupArtifact{}

func NewGroup (
	key abs.ArtifactKey,
	name string,
	arise *abs.AriseRef,
) *GroupArtifact {
	artifact := &GroupArtifact {}
	artifact.Key = key
	artifact.ID = abs.NextArtifactID()
	artifact.Name = name
	artifact.Arise = arise
	return artifact
}

// ---------------------------------------- Step ----------------------------------------

type StepBase struct {
	Description string
}

func (step *StepBase) SimpleDescr() string {
	return step.Description
}

// ---------------------------------------- Transform ----------------------------------------

type TransformBase struct {
	Description string
	Arise *abs.AriseRef
}

func (base *TransformBase) TransformDescr() string {
	return base.Description
}

func (base *TransformBase) TransformArise() *abs.AriseRef {
	return base.Arise
}

func PlanSingleTransform(
	transform abs.Transform,
	source, destination abs.Artifact,
	plan *abs.Plan,
	planner func() abs.BuildError,
) abs.BuildError {
	smod, serr, _ := source.LatestModTime()
	if serr != nil {
		serr.AddErrorFrame(&ApplyTransformFrame {
			Transform: transform,
		})
		return serr
	}
	dmod, derr, dmiss := destination.EarliestModTime()
	if derr != nil {
		derr.AddErrorFrame(&ApplyTransformFrame {
			Transform: transform,
		})
		return derr
	}
	if dmiss || smod.After(dmod) {
		perr := planner()
		if perr != nil {
			return perr
		}
	}
	return nil
}

func PlanMultiTransform(
	transform abs.Transform,
	sources []abs.Artifact,
	destination abs.Artifact,
	plan *abs.Plan,
	planner func() abs.BuildError,
) abs.BuildError {
	dmod, derr, dmiss := destination.EarliestModTime()
	if derr != nil {
		derr.AddErrorFrame(&ApplyTransformFrame {
			Transform: transform,
		})
		return derr
	}
	apply := dmiss
	for _, source := range sources {
		if apply {
			break
		}
		smod, serr, _ := source.LatestModTime()
		if serr != nil {
			serr.AddErrorFrame(&ApplyTransformFrame {
				Transform: transform,
			})
			return serr
		}
		if smod.After(dmod) {
			apply = true
		}
	}
	if apply {
		perr := planner()
		if perr != nil {
			return perr
		}
	}
	return nil
}

type SingleTransformBase struct {
	TransformBase
	Source abs.Artifact
}

type MultiTransformBase struct {
	TransformBase
	Sources []abs.Artifact
}

func (base *MultiTransformBase) AddSource(source abs.Artifact) {
	base.Sources = append(base.Sources, source)
}

// ---------------------------------------- Action ----------------------------------------

type ActionBase struct {
	Arise *abs.AriseRef
}

func (base *ActionBase) ActionArise() *abs.AriseRef {
	return base.Arise
}

type AttainAction struct {
	ActionBase
	Goal *abs.Goal
}

func (action *AttainAction) SimpleDescr() string {
	return "attain " + action.Goal.Name
}

func (action *AttainAction) Perform(plan *abs.Plan) abs.BuildError {
	return Attain(action.Goal, plan)
}

var _ abs.Action = &AttainAction{}

type RequireAction struct {
	ActionBase
	Artifact abs.Artifact
}

func (action *RequireAction) SimpleDescr() string {
	return "require " + action.Artifact.DisplayName()
}

func (action *RequireAction) Perform(plan *abs.Plan) abs.BuildError {
	return action.Artifact.Require(plan)
}

var _ abs.Action = &RequireAction{}

// ---------------------------------------- Goal ----------------------------------------

func Attain(goal *abs.Goal, plan *abs.Plan) (err abs.BuildError) {
	for _, action := range goal.Actions() {
		err = action.Perform(plan)
		if err != nil {
			err.AddErrorFrame(&AttainGoalFrame {
				Goal: goal,
			})
			return
		}
	}
	return
}
