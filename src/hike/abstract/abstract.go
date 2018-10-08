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

func (ref *AriseRef) PrintArise(level uint) (err error) {
	_, err = fmt.Fprintln(os.Stderr, "arising from")
	if err != nil {
		return
	}
	err = IndentError(level + 1)
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(os.Stderr, ref.Text)
	if err != nil {
		return
	}
	err = IndentError(level)
	if err != nil {
		return
	}
	location, err := ref.Location.Format()
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(os.Stderr, "at %s", location)
	return
}

type BuildFrame interface {
	PrintErrorFrame(level uint) error
}

type BuildError interface {
	PrintBuildError(level uint) error
	AddErrorFrame(frame BuildFrame)
}

type ArtifactKey struct {
	Project string
	Artifact string
}

func (key *ArtifactKey) Unified() string {
	return key.Project + "::" + key.Artifact
}

type Artifact interface {
	ArtifactKey() *ArtifactKey
	DisplayName() string
	ArtifactArise() *AriseRef
	PathNames(sink []string) []string
	EarliestModTime() (time.Time, BuildError, bool)
	LatestModTime() (time.Time, BuildError, bool)
	Flatten() BuildError
	Require(plan *Plan) BuildError
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

type Config struct {
	ProjectName string
	InducedProjectName string
	TopDir string
}

func (config *Config) EffectiveProjectName() string {
	switch {
		case len(config.InducedProjectName) > 0:
			return config.InducedProjectName
		case len(config.ProjectName) > 0:
			return config.ProjectName
		default:
			return "this"
	}
}

type SpecState struct {
	Config *Config
	PipelineTip *Artifact
}

type Plan struct {
	steps []Step
	knownUpToDate map[*Artifact]bool
}

func (plan *Plan) AddStep(step Step) {
	plan.steps = append(plan.steps, step)
}

func (plan *Plan) StepCount() int {
	return len(plan.steps)
}

func (plan *Plan) BroughtUpToDate(artifact *Artifact) {
	plan.knownUpToDate[artifact] = true
}

func (plan *Plan) AlreadyUpToDate(artifact *Artifact) bool {
	return plan.knownUpToDate[artifact]
}
