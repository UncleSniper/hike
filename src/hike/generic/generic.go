package generic

import (
	abs "hike/abstract"
	con "hike/concrete"
)

// ---------------------------------------- Step ----------------------------------------

type VariableCommandLine func([]string, []string) [][]string

type CommandStep struct {
	con.StepBase
	Sources []abs.Artifact
	Destination []abs.Artifact
	CommandLine VariableCommandLine
}

func (step *CommandStep) Perform() abs.BuildError {
	//TODO
	return nil
}

var _ abs.Step = &CommandStep{}

// ---------------------------------------- Transform ----------------------------------------
