package spec

import (
	abs "hike/abstract"
	con "hike/concrete"
	loc "hike/location"
)

type DuplicateGoalError struct {
	con.BuildErrorBase
	RegisterArise *abs.AriseRef
	OldGoal *abs.Goal
	NewGoal *abs.Goal
}

func (duplicate *DuplicateGoalError) PrintBuildError(level uint) error {
	prn := &abs.ErrorPrinter{}
	prn.Level(level)
	prn.Println("Goal name clash:", duplicate.OldGoal.Name)
	prn.Indent(1)
	prn.Arise(duplicate.RegisterArise, 1)
	prn.Println()
	prn.Indent(1)
	prn.Print("between old goal ")
	prn.Arise(duplicate.OldGoal.Arise, 1)
	prn.Println()
	prn.Indent(1)
	prn.Print("and new goal ")
	prn.Arise(duplicate.NewGoal.Arise, 1)
	duplicate.InjectBacktrace(prn, 0)
	return prn.Done()
}

var _ abs.BuildError = &DuplicateGoalError{}

type NoSuchGoalError struct {
	con.BuildErrorBase
	Name string
	ReferenceLocation *loc.Location
	ReferenceArise *abs.AriseRef
}

func (no *NoSuchGoalError) PrintBuildError(level uint) error {
	prn := &abs.ErrorPrinter{}
	prn.Level(level)
	prn.Println("No such goal:", no.Name)
	prn.Indent(1)
	prn.Print("referenced at ")
	prn.Location(no.ReferenceLocation)
	prn.Println()
	prn.Indent(1)
	prn.Print("reference ")
	prn.Arise(no.ReferenceArise, 1)
	no.InjectBacktrace(prn, 0)
	return prn.Done()
}

var _ abs.BuildError = &NoSuchGoalError{}

type DuplicateArtifactError struct {
	con.BuildErrorBase
	RegisterArise *abs.AriseRef
	OldArtifact abs.Artifact
	NewArtifact abs.Artifact
}

func (duplicate *DuplicateArtifactError) PrintBuildError(level uint) error {
	prn := &abs.ErrorPrinter{}
	prn.Level(level)
	prn.Println("Artifact key clash:", duplicate.OldArtifact.ArtifactKey().Unified())
	prn.Indent(1)
	prn.Arise(duplicate.RegisterArise, 1)
	prn.Println()
	prn.Indent(1)
	prn.Print("between old artifact ")
	prn.Arise(duplicate.OldArtifact.ArtifactArise(), 1)
	prn.Println()
	prn.Indent(1)
	prn.Print("and new artifact ")
	prn.Arise(duplicate.NewArtifact.ArtifactArise(), 1)
	duplicate.InjectBacktrace(prn, 0)
	return prn.Done()
}

var _ abs.BuildError = &DuplicateArtifactError{}

type NoSuchArtifactError struct {
	con.BuildErrorBase
	Key *abs.ArtifactKey
	ReferenceLocation *loc.Location
	ReferenceArise *abs.AriseRef
}

func (no *NoSuchArtifactError) PrintBuildError(level uint) error {
	prn := &abs.ErrorPrinter{}
	prn.Level(level)
	prn.Println("No such artifact:", no.Key.Unified())
	prn.Indent(1)
	prn.Print("referenced at ")
	prn.Location(no.ReferenceLocation)
	prn.Println()
	prn.Indent(1)
	prn.Print("reference ")
	prn.Arise(no.ReferenceArise, 1)
	no.InjectBacktrace(prn, 0)
	return prn.Done()
}

var _ abs.BuildError = &NoSuchArtifactError{}

type PendingResolver func() abs.BuildError

type State struct {
	Config *Config
	PipelineTip abs.Artifact
	goals map[string]*abs.Goal
	artifacts map[string]abs.Artifact
	pendingResolutions []PendingResolver
}

func NewState(config *Config) *State {
	return &State {
		Config: config,
		goals: make(map[string]*abs.Goal),
		artifacts: make(map[string]abs.Artifact),
	}
}

func (state *State) Goal(name string) *abs.Goal {
	return state.goals[name]
}

func (state *State) RegisterGoal(goal *abs.Goal, arise *abs.AriseRef) *DuplicateGoalError {
	old, present := state.goals[goal.Name]
	if present {
		return &DuplicateGoalError {
			RegisterArise: arise,
			OldGoal: old,
			NewGoal: goal,
		}
	}
	state.goals[goal.Name] = goal
	return nil
}

func (state *State) Artifact(key *abs.ArtifactKey) abs.Artifact {
	return state.artifacts[key.Unified()]
}

func (state *State) RegisterArtifact(artifact abs.Artifact, arise *abs.AriseRef) *DuplicateArtifactError {
	ks := artifact.ArtifactKey().Unified()
	old, present := state.artifacts[ks]
	if present {
		return &DuplicateArtifactError {
			RegisterArise: arise,
			OldArtifact: old,
			NewArtifact: artifact,
		}
	}
	state.artifacts[ks] = artifact
	return nil
}

func (state *State) SlateResolver(resolver PendingResolver) {
	state.pendingResolutions = append(state.pendingResolutions, resolver)
}

func (state *State) FlushPendingResolutions() abs.BuildError {
	for {
		if len(state.pendingResolutions) == 0 {
			return nil
		}
		pr := state.pendingResolutions
		state.pendingResolutions = nil
		for _, resolver := range pr {
			err := resolver()
			if err != nil {
				return err
			}
		}
	}
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
