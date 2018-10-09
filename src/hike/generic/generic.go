package generic

import (
	"os"
	"fmt"
	"bufio"
	"os/exec"
	"strings"
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
	con.BuildErrorBase
	Argv []string
	Fault error
	Output []byte
}

func (failed *CommandFailedError) PrintBuildError(level uint) (err error) {
	_, err = fmt.Fprintln(os.Stderr, "Command")
	if err != nil {
		return
	}
	err = abs.IndentError(level + 1)
	if err != nil {
		return
	}
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
		_, err = fmt.Fprintf(os.Stderr, "%s%s%s%s", sep, delim, word, delim)
		if err != nil {
			return
		}
	}
	_, err = fmt.Fprintln(os.Stderr)
	if err != nil {
		return
	}
	err = abs.IndentError(level)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(os.Stderr, "failed: %s", failed.Fault.Error())
	if err != nil {
		return
	}
	if len(failed.Output) > 0 {
		_, err = fmt.Fprintln(os.Stderr)
		if err != nil {
			return
		}
		err = abs.IndentError(level)
		if err != nil {
			return
		}
		_, err = fmt.Fprint(os.Stderr, "Output:")
		if err != nil {
			return
		}
		sout := string(failed.Output)
		sread := strings.NewReader(sout)
		scan := bufio.NewScanner(sread)
		for scan.Scan() {
			_, err = fmt.Fprintln(os.Stderr)
			if err != nil {
				return
			}
			err = abs.IndentError(level + 1)
			if err != nil {
				return
			}
			_, err = fmt.Fprint(os.Stderr, scan.Text())
			if err != nil {
				return
			}
		}
		err = scan.Err()
		if err != nil {
			return
		}
	}
	err = failed.PrintBacktrace(level)
	return
}

var _ abs.BuildError = &CommandFailedError{}

// ---------------------------------------- Step ----------------------------------------

type VariableCommandLine func([]string, []string) [][]string

type CommandStep struct {
	con.StepBase
	Sources []abs.Artifact
	Destination abs.Artifact
	CommandLine VariableCommandLine
	Loud bool
}

func (step *CommandStep) Perform() abs.BuildError {
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
) {
	step := &CommandStep {
		Sources: sources,
		Destination: destination,
		CommandLine: base.CommandLine,
		Loud: base.Loud,
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

func (transform *SingleCommandTransform) Plan(destination abs.Artifact, plan *abs.Plan) abs.BuildError {
	return con.PlanSingleTransform(transform, transform.Source, destination, plan, func() abs.BuildError {
		transform.PlanCommandTransform(transform.Description, []abs.Artifact{transform.Source}, destination, plan)
		return nil
	})
}

var _ abs.Transform = &SingleCommandTransform{}

type MultiCommandTransform struct {
	con.MultiTransformBase
	CommandTransformBase
}

func (transform *MultiCommandTransform) Plan(destination abs.Artifact, plan *abs.Plan) abs.BuildError {
	return con.PlanMultiTransform(transform, transform.Sources, destination, plan, func() abs.BuildError {
		transform.PlanCommandTransform(transform.Description, transform.Sources, destination, plan)
		return nil
	})
}

var _ abs.Transform = &MultiCommandTransform{}
