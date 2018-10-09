package spec

import (
	"os"
	"fmt"
	abs "hike/abstract"
	con "hike/concrete"
)

type DuplicateGoalError struct {
	con.BuildErrorBase
	RegisterArise *abs.AriseRef
	OldGoal abs.Goal
	NewGoal abs.Goal
}

func (duplicate *DuplicateGoalError) PrintBuildError(level uint) (err error) {
	_, err = fmt.Fprintln(os.Stderr, "Goal name clash:", duplicate.OldGoal.Name)
	if err != nil {
		return
	}
	err = abs.IndentError(level)
	if err != nil {
		return
	}
	_, err = fmt.Fprint(os.Stderr, "between old goal ")
	if err != nil {
		return
	}
	err = duplicate.OldGoal.Arise.PrintArise(level)
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(os.Stderr)
	if err != nil {
		return
	}
	err = abs.IndentError(level)
	if err != nil {
		return
	}
	_, err = fmt.Fprint(os.Stderr, "and new goal ")
	if err != nil {
		return
	}
	err = duplicate.NewGoal.Arise.PrintArise(level)
	if err != nil {
		return
	}
	err = duplicate.PrintBacktrace(level)
	return
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
