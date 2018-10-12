package abstract

import (
	"time"
	herr "hike/error"
)

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
	ArtifactArise() *herr.AriseRef
	PathNames(sink []string) []string
	EarliestModTime(arise *herr.AriseRef) (time.Time, herr.BuildError, bool)
	LatestModTime(arise *herr.AriseRef) (time.Time, herr.BuildError, bool)
	Flatten() herr.BuildError
	Require(plan *Plan, arise *herr.AriseRef) herr.BuildError
}

var nextArtifactID uint = 0

func NextArtifactID() ArtifactID {
	id := ArtifactID(nextArtifactID)
	nextArtifactID++
	return id
}

type Step interface {
	Perform() herr.BuildError
	SimpleDescr() string
}

type Transform interface {
	TransformDescr() string
	TransformArise() *herr.AriseRef
	Plan(destination Artifact, plan *Plan) herr.BuildError
}

type Action interface {
	SimpleDescr() string
	Perform(plan *Plan) herr.BuildError
	ActionArise() *herr.AriseRef
}

type Goal struct {
	Name string
	Label string
	Arise *herr.AriseRef
	actions []Action
}

func (goal *Goal) AddAction(action Action) {
	goal.actions = append(goal.actions, action)
}

func (goal *Goal) Actions() []Action {
	return goal.actions
}

func (goal *Goal) ActionCount() int {
	return len(goal.actions)
}

type Plan struct {
	steps []Step
	knownUpToDate map[ArtifactID]bool
}

func NewPlan() *Plan {
	return &Plan {
		knownUpToDate: make(map[ArtifactID]bool),
	}
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
