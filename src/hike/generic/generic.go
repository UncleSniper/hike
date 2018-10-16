package generic

import (
	"fmt"
	"bufio"
	"os/exec"
	"strings"
	herr "hike/error"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
)

// ---------------------------------------- Artifact ----------------------------------------

func PathsOfArtifacts(artifacts []abs.Artifact) []string {
	var all []string
	for _, artifact := range artifacts {
		all = artifact.PathNames(all)
	}
	return all
}

// ---------------------------------------- BuildError ----------------------------------------

type CommandFailedError struct {
	herr.BuildErrorBase
	Argv []string
	Fault error
	Output []byte
	ExecArise *herr.AriseRef
}

func (failed *CommandFailedError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Println("Command")
	prn.Indent(1)
	first := true
	var sep, delim string
	for _, word := range failed.Argv {
		if first {
			first = false
			sep = ""
		} else {
			sep = " "
		}
		if strings.IndexRune(word, ' ') < 0 {
			delim = ""
		} else {
			delim = "'"
		}
		prn.Printf("%s%s%s%s", sep, delim, word, delim)
	}
	prn.Println()
	prn.Indent(0)
	prn.Printf("failed: %s\n", failed.Fault.Error())
	prn.Indent(0)
	prn.Print("during execution ")
	prn.Arise(failed.ExecArise, 0)
	if len(failed.Output) > 0 {
		prn.Println()
		prn.Indent(0)
		prn.Print("Output:")
		sout := string(failed.Output)
		sread := strings.NewReader(sout)
		scan := bufio.NewScanner(sread)
		for scan.Scan() {
			prn.Println()
			prn.Indent(1)
			prn.Print(scan.Text())
		}
		err := scan.Err()
		if err != nil {
			return err
		}
	}
	failed.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (failed *CommandFailedError) BuildErrorLocation() *loc.Location {
	return failed.ExecArise.Location
}

var _ herr.BuildError = &CommandFailedError{}

// ---------------------------------------- Step ----------------------------------------

type VariableCommandLine func([]string, []string) [][]string

type CommandStep struct {
	con.StepBase
	Sources []abs.Artifact
	Destination abs.Artifact
	CommandLine VariableCommandLine
	Loud bool
	CommandArise *herr.AriseRef
}

func (step *CommandStep) Perform() herr.BuildError {
	argvs := step.CommandLine(PathsOfArtifacts(step.Sources), step.Destination.PathNames(nil))
	for _, argv := range argvs {
		if len(argv) == 0 {
			continue
		}
		cmd := exec.Command(argv[0])
		cmd.Args = argv
		out, err := cmd.CombinedOutput()
		if err != nil {
			return &CommandFailedError {
				Argv: argv,
				Fault: err,
				Output: out,
				ExecArise: step.CommandArise,
			}
		}
		if step.Loud {
			fmt.Print(string(out))
		}
	}
	return nil
}

var _ abs.Step = &CommandStep{}

// ---------------------------------------- Transform ----------------------------------------

type CommandTransformBase struct {
	CommandLine VariableCommandLine
	Loud bool
	SuffixIsDestination bool
}

func (base *CommandTransformBase) PlanCommandTransform(
	descriptionPrefix string,
	sources []abs.Artifact,
	destination abs.Artifact,
	plan *abs.Plan,
	transformArise *herr.AriseRef,
) {
	step := &CommandStep {
		Sources: sources,
		Destination: destination,
		CommandLine: base.CommandLine,
		Loud: base.Loud,
		CommandArise: transformArise,
	}
	var suffix string
	if base.SuffixIsDestination || len(sources) != 1 {
		suffix = destination.DisplayName()
	} else {
		suffix = sources[0].DisplayName()
	}
	step.Description = fmt.Sprintf("%s %s", descriptionPrefix, suffix)
	plan.AddStep(step)
}

type SingleCommandTransform struct {
	con.SingleTransformBase
	CommandTransformBase
}

func (transform *SingleCommandTransform) Plan(destination abs.Artifact, plan *abs.Plan) herr.BuildError {
	return con.PlanSingleTransform(transform, transform.Source, destination, plan, func() herr.BuildError {
		transform.PlanCommandTransform(
			transform.Description,
			[]abs.Artifact{transform.Source},
			destination,
			plan,
			transform.TransformArise(),
		)
		return nil
	})
}

var _ abs.Transform = &SingleCommandTransform{}

func NewSingleCommandTransform(
	description string,
	arise *herr.AriseRef,
	source abs.Artifact,
	commandLine VariableCommandLine,
	loud bool,
	suffixIsDestination bool,
) *SingleCommandTransform {
	transform := &SingleCommandTransform {}
	transform.Description = description
	transform.Arise = arise
	transform.Source = source
	transform.CommandLine = commandLine
	transform.Loud = loud
	transform.SuffixIsDestination = suffixIsDestination
	return transform
}

type MultiCommandTransform struct {
	con.MultiTransformBase
	CommandTransformBase
}

func (transform *MultiCommandTransform) Plan(destination abs.Artifact, plan *abs.Plan) herr.BuildError {
	return con.PlanMultiTransform(transform, transform.Sources, destination, plan, func() herr.BuildError {
		transform.PlanCommandTransform(
			transform.Description,
			transform.Sources,
			destination,
			plan,
			transform.TransformArise(),
		)
		return nil
	})
}

var _ abs.Transform = &MultiCommandTransform{}

func NewMultiCommandTransform(
	description string,
	arise *herr.AriseRef,
	commandLine VariableCommandLine,
	loud bool,
	suffixIsDestination bool,
) *MultiCommandTransform {
	transform := &MultiCommandTransform {}
	transform.Description = description
	transform.Arise = arise
	transform.CommandLine = commandLine
	transform.Loud = loud
	transform.SuffixIsDestination = suffixIsDestination
	return transform
}
