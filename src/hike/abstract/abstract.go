package abstract

import (
	"os"
	"fmt"
	"time"
	loc "hike/location"
)

func IndentError(level uint) (err error) {
	for ; level > 0; level-- {
		_, err = fmt.Fprint(os.Stderr, "    ")
		if err != nil {
			break
		}
	}
	return
}

type AriseRef struct {
	Text string
	Location *loc.Location
}

func (ref *AriseRef) PrintArise(level uint) error {
	prn := &ErrorPrinter{}
	prn.Println("arising from")
	prn.Indent(level + 1)
	prn.Println(ref.Text)
	prn.Indent(level)
	location, err := ref.Location.Format()
	if err != nil {
		return err
	}
	prn.Printf("at %s", location)
	return prn.Done()
}

type BuildFrame interface {
	PrintErrorFrame(level uint) error
}

type BuildError interface {
	PrintBuildError(level uint) error
	AddErrorFrame(frame BuildFrame)
}

type ErrorPrinter struct {
	firstError error
	level uint
}

func (printer *ErrorPrinter) Print(values ...interface{}) {
	if printer.firstError == nil {
		_, printer.firstError = fmt.Fprint(os.Stderr, values...)
	}
}

func (printer *ErrorPrinter) Println(values ...interface{}) {
	if printer.firstError == nil {
		_, printer.firstError = fmt.Fprintln(os.Stderr, values...)
	}
}

func (printer *ErrorPrinter) Printf(format string, values ...interface{}) {
	if printer.firstError == nil {
		_, printer.firstError = fmt.Fprintf(os.Stderr, format, values...)
	}
}

func (printer *ErrorPrinter) Level(level uint) {
	printer.level = level
}

func (printer *ErrorPrinter) Indent(level uint) {
	if printer.firstError == nil {
		printer.firstError = IndentError(printer.level + level)
	}
}

func (printer *ErrorPrinter) Arise(arise *AriseRef, level uint) {
	if printer.firstError == nil {
		printer.firstError = arise.PrintArise(printer.level + level)
	}
}

func (printer *ErrorPrinter) Frame(frame BuildFrame, level uint) {
	if printer.firstError == nil {
		printer.firstError = frame.PrintErrorFrame(printer.level + level)
	}
}

func (printer *ErrorPrinter) Inject(callback func(level uint) error, level uint) {
	if printer.firstError == nil {
		printer.firstError = callback(printer.level + level)
	}
}

func (printer *ErrorPrinter) Done() error {
	return printer.firstError
}

type ArtifactKey struct {
	Project string
	Artifact string
}

func (key *ArtifactKey) Unified() string {
	return key.Project + "::" + key.Artifact
}

type ArtifactID uint

type Artifact interface {
	ArtifactKey() *ArtifactKey
	ArtifactID() ArtifactID
	DisplayName() string
	ArtifactArise() *AriseRef
	PathNames(sink []string) []string
	EarliestModTime() (time.Time, BuildError, bool)
	LatestModTime() (time.Time, BuildError, bool)
	Flatten() BuildError
	Require(plan *Plan) BuildError
}

var nextArtifactID uint = 0

func NextArtifactID() ArtifactID {
	id := ArtifactID(nextArtifactID)
	nextArtifactID++
	return id
}

type Step interface {
	Perform() BuildError
	SimpleDescr() string
}

type Transform interface {
	TransformDescr() string
	TransformArise() *AriseRef
	Plan(destination Artifact, plan *Plan) BuildError
}

type Action interface {
	SimpleDescr() string
	Perform(plan *Plan) BuildError
	ActionArise() *AriseRef
}

type Goal struct {
	Name string
	Arise *AriseRef
	actions []Action
}

func (goal *Goal) AddAction(action Action) {
	goal.actions = append(goal.actions, action)
}

func (goal *Goal) Actions() []Action {
	return goal.actions
}

type Plan struct {
	steps []Step
	knownUpToDate map[ArtifactID]bool
}

func (plan *Plan) AddStep(step Step) {
	plan.steps = append(plan.steps, step)
}

func (plan *Plan) StepCount() int {
	return len(plan.steps)
}

func (plan *Plan) BroughtUpToDate(artifact Artifact) {
	plan.knownUpToDate[artifact.ArtifactID()] = true
}

func (plan *Plan) AlreadyUpToDate(artifact Artifact) bool {
	return plan.knownUpToDate[artifact.ArtifactID()]
}
