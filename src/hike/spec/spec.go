package spec

import (
	abs "hike/abstract"
	con "hike/concrete"
)

type DuplicateGoalError struct {
	con.BuildErrorBase
	RegisterArise *abs.AriseRef
	OldGoal abs.Goal
	NewGoal abs.Goal
}

func (duplicate *DuplicateGoalError) PrintBuildError(level uint) error {
	prn := &abs.ErrorPrinter{}
	prn.Level(level)
	prn.Println("Goal name clash:", duplicate.OldGoal.Name)
	prn.Indent(0)
	prn.Print("between old goal ")
	prn.Arise(duplicate.OldGoal.Arise, 0)
	prn.Println()
	prn.Indent(0)
	prn.Print("and new goal ")
	prn.Arise(duplicate.NewGoal.Arise, 0)
	duplicate.InjectBacktrace(prn, 0)
	return prn.Done()
}

var _ abs.BuildError = &DuplicateGoalError{}

type State struct {
	Config *Config
	PipelineTip *abs.Artifact
	goals map[string]abs.Goal
}

func (state *State) Goal(name string) abs.Goal {
	return state.goals[name]
}

func (state *State) RegisterGoal(goal abs.Goal, arise *abs.AriseRef) *DuplicateGoalError {
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
