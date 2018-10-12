package concrete

import (
	"os"
	"time"
	herr "hike/error"
	loc "hike/location"
	abs "hike/abstract"
)

// ---------------------------------------- BuildFrame ----------------------------------------

func PrintArtifactErrorFrameBase(level uint, action string, artifact abs.Artifact) error {
	prn := &herr.ErrorPrinter{}
	prn.Level(level)
	prn.Printf("%s artifact\n", action)
	prn.Indent(1)
	prn.Printf("%s [%s]\n", artifact.DisplayName(), artifact.ArtifactKey().Unified())
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

var _ herr.BuildFrame = &RequireArtifactFrame{}

type FlattenArtifactFrame struct {
	Artifact abs.Artifact
}

func (frame *FlattenArtifactFrame) PrintErrorFrame(level uint) error {
	return PrintArtifactErrorFrameBase(level, "flattening", frame.Artifact)
}

var _ herr.BuildFrame = &FlattenArtifactFrame{}

type ApplyTransformFrame struct {
	Transform abs.Transform
}

func (frame *ApplyTransformFrame) PrintErrorFrame(level uint) error {
	prn := &herr.ErrorPrinter{}
	prn.Level(level)
	prn.Println("applying transform")
	prn.Indent(1)
	prn.Println(frame.Transform.TransformDescr())
	prn.Indent(0)
	prn.Arise(frame.Transform.TransformArise(), 0)
	return prn.Done()
}

var _ herr.BuildFrame = &ApplyTransformFrame{}

type AttainGoalFrame struct {
	Goal *abs.Goal
}

func (frame *AttainGoalFrame) PrintErrorFrame(level uint) error {
	prn := &herr.ErrorPrinter{}
	prn.Printf("attaining goal '%s' ", frame.Goal.Name)
	prn.Arise(frame.Goal.Arise, level)
	return prn.Done()
}

var _ herr.BuildFrame = &AttainGoalFrame{}

type PerformActionFrame struct {
	Action abs.Action
}

func (frame *PerformActionFrame) PrintErrorFrame(level uint) error {
	prn := &herr.ErrorPrinter{}
	prn.Printf("performing action '%s' ", frame.Action.SimpleDescr())
	prn.Arise(frame.Action.ActionArise(), level)
	return prn.Done()
}

// ---------------------------------------- BuildError ----------------------------------------

type NoGeneratorError struct {
	herr.BuildErrorBase
	Artifact abs.Artifact
	RequireArise *herr.AriseRef
}

func (nogen *NoGeneratorError) PrintBuildError(level uint) error {
	prn := &herr.ErrorPrinter{}
	prn.Level(level)
	prn.Println("Don't know how to obtain artifact")
	prn.Indent(1)
	prn.Printf("%s [%s]\n", nogen.Artifact.DisplayName(), nogen.Artifact.ArtifactKey().Unified())
	prn.Indent(0)
	prn.Arise(nogen.Artifact.ArtifactArise(), 0)
	prn.Println()
	prn.Indent(0)
	prn.Print("for requisition ")
	prn.Arise(nogen.RequireArise, 0)
	nogen.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (nogen *NoGeneratorError) BuildErrorLocation() *loc.Location {
	return nogen.RequireArise.Location
}

var _ herr.BuildError = &NoGeneratorError{}

type CannotStatError struct {
	herr.BuildErrorBase
	Path string
	OSError error
	OperationArise *herr.AriseRef
}

func (cannot *CannotStatError) PrintBuildError(level uint) error {
	prn := &herr.ErrorPrinter{}
	prn.Level(level)
	prn.Println("Failed to stat file")
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

func (cannot *CannotStatError) BuildErrorLocation() *loc.Location {
	return cannot.OperationArise.Location
}

var _ herr.BuildError = &CannotStatError{}

// ---------------------------------------- Artifact ----------------------------------------

type ArtifactBase struct {
	Key abs.ArtifactKey
	ID abs.ArtifactID
	Name string
	Arise *herr.AriseRef
}

func (artifact *ArtifactBase) ArtifactKey() *abs.ArtifactKey {
	return &artifact.Key
}

func (artifact *ArtifactBase) ArtifactID() abs.ArtifactID {
	return artifact.ID
}

func (artifact *ArtifactBase) ArtifactArise() *herr.AriseRef {
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

func (artifact *FileArtifact) ModTime(arise *herr.AriseRef) (stamp time.Time, err herr.BuildError, missing bool) {
	info, oserr := os.Stat(artifact.Path)
	if oserr == nil {
		stamp = info.ModTime()
	} else {
		missing = os.IsNotExist(oserr)
		if !missing {
			err = &CannotStatError {
				Path: artifact.Path,
				OSError: oserr,
				OperationArise: arise,
			}
		}
	}
	return
}

func (artifact *FileArtifact) EarliestModTime(arise *herr.AriseRef) (time.Time, herr.BuildError, bool) {
	return artifact.ModTime(arise)
}

func (artifact *FileArtifact) LatestModTime(arise *herr.AriseRef) (time.Time, herr.BuildError, bool) {
	return artifact.ModTime(arise)
}

func (artifact *FileArtifact) Flatten() herr.BuildError {
	return nil
}

func FileExists(path string, arise *herr.AriseRef) (exists bool, err herr.BuildError) {
	_, oserr := os.Lstat(path)
	switch {
		case oserr == nil:
			exists = true
		case !os.IsNotExist(oserr):
			err = &CannotStatError {
				Path: path,
				OSError: oserr,
				OperationArise: arise,
			}
	}
	return
}

func (artifact *FileArtifact) Require(plan *abs.Plan, requireArise *herr.AriseRef) (err herr.BuildError) {
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
		exists, err := FileExists(artifact.Path, requireArise)
		switch {
			case err != nil:
				err.AddErrorFrame(&RequireArtifactFrame {
					Artifact: artifact,
				})
			case !exists:
				err = &NoGeneratorError {
					Artifact: artifact,
					RequireArise: requireArise,
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
	arise *herr.AriseRef,
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

func (artifact *GroupArtifact) EarliestModTime(arise *herr.AriseRef) (
	result time.Time,
	err herr.BuildError,
	missing bool,
) {
	result = time.Now()
	var (
		have bool
		chmod time.Time
		chmiss bool
	)
	for _, child := range artifact.children {
		chmod, err, chmiss = child.EarliestModTime(arise)
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

func (artifact *GroupArtifact) LatestModTime(arise *herr.AriseRef) (
	result time.Time,
	err herr.BuildError,
	missing bool,
) {
	result = time.Now()
	var (
		have bool
		chmod time.Time
		chmiss bool
	)
	for _, child := range artifact.children {
		chmod, err, chmiss = child.LatestModTime(arise)
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

func (artifact *GroupArtifact) Flatten() (err herr.BuildError) {
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

func (artifact *GroupArtifact) Require(plan *abs.Plan, requireArise *herr.AriseRef) (err herr.BuildError) {
	if plan.AlreadyUpToDate(artifact) {
		return
	}
	for _, child := range artifact.children {
		err = child.Require(plan, requireArise)
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
	arise *herr.AriseRef,
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
	Arise *herr.AriseRef
}

func (base *TransformBase) TransformDescr() string {
	return base.Description
}

func (base *TransformBase) TransformArise() *herr.AriseRef {
	return base.Arise
}

func PlanSingleTransform(
	transform abs.Transform,
	source, destination abs.Artifact,
	plan *abs.Plan,
	planner func() herr.BuildError,
) herr.BuildError {
	smod, serr, _ := source.LatestModTime(transform.TransformArise())
	if serr != nil {
		serr.AddErrorFrame(&ApplyTransformFrame {
			Transform: transform,
		})
		return serr
	}
	dmod, derr, dmiss := destination.EarliestModTime(transform.TransformArise())
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
	planner func() herr.BuildError,
) herr.BuildError {
	dmod, derr, dmiss := destination.EarliestModTime(transform.TransformArise())
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
		smod, serr, _ := source.LatestModTime(transform.TransformArise())
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

func (base *MultiTransformBase) SourceCount() int {
	return len(base.Sources)
}

// ---------------------------------------- Action ----------------------------------------

type ActionBase struct {
	Arise *herr.AriseRef
}

func (base *ActionBase) ActionArise() *herr.AriseRef {
	return base.Arise
}

type AttainAction struct {
	ActionBase
	Goal *abs.Goal
}

func (action *AttainAction) SimpleDescr() string {
	return "attain " + action.Goal.Name
}

func (action *AttainAction) Perform(plan *abs.Plan) herr.BuildError {
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

func (action *RequireAction) Perform(plan *abs.Plan) herr.BuildError {
	return action.Artifact.Require(plan, action.ActionArise())
}

var _ abs.Action = &RequireAction{}

// ---------------------------------------- Goal ----------------------------------------

func Attain(goal *abs.Goal, plan *abs.Plan) (err herr.BuildError) {
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
