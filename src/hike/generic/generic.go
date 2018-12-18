package generic

import (
	"os"
	"fmt"
	"bufio"
	"os/exec"
	"strings"
	herr "hike/error"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
)

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

type VariableCommandLine func([]string, []string) ([][]string, herr.BuildError)
type CommandLineDumper func(uint) error
type CommandWordsRequirer func(*abs.Plan, *herr.AriseRef) herr.BuildError

type CommandStep struct {
	con.StepBase
	Sources []abs.Artifact
	Destination abs.Artifact
	CommandLine VariableCommandLine
	Loud bool
	CommandArise *herr.AriseRef
}

func (step *CommandStep) Perform() herr.BuildError {
	destPaths, err := step.Destination.PathNames(nil)
	if err != nil {
		return err
	}
	for _, destDir := range destPaths {
		err = con.MakeEnclosingDirectories(destDir, step.CommandArise)
		if err != nil {
			return err
		}
	}
	srcPaths, err := con.PathsOfArtifacts(step.Sources)
	if err != nil {
		return err
	}
	argvs, err := step.CommandLine(srcPaths, destPaths)
	if err != nil {
		return err
	}
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

type DeletePathStep struct {
	con.StepBase
	Path string
	DeleteArise *herr.AriseRef
}

func (step *DeletePathStep) Perform() herr.BuildError {
	nerr := os.RemoveAll(step.Path)
	if nerr == nil {
		return nil
	}
	return &con.CannotDeleteFileError {
		Path: step.Path,
		OSError: nerr,
		OperationArise: step.DeleteArise,
	}
}

var _ abs.Step = &DeletePathStep{}

type DeleteArtifactStep struct {
	con.StepBase
	Artifact abs.Artifact
	DeleteArise *herr.AriseRef
}

func (step *DeleteArtifactStep) Perform() herr.BuildError {
	paths, err := step.Artifact.PathNames(nil)
	if err != nil {
		return err
	}
	for _, path := range paths {
		nerr := os.RemoveAll(path)
		if nerr != nil {
			return &con.CannotDeleteFileError {
				Path: path,
				OSError: nerr,
				OperationArise: step.DeleteArise,
			}
		}
	}
	return nil
}

var _ abs.Step = &DeleteArtifactStep{}

// ---------------------------------------- Transform ----------------------------------------

type CommandTransformBase struct {
	CommandLine VariableCommandLine
	DumpCommandLine CommandLineDumper
	RequireCommandWords CommandWordsRequirer
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
	step.Description = fmt.Sprintf("[%s] %s %s", destination.ArtifactKey().Project, descriptionPrefix, suffix)
	plan.AddStep(step)
}

func (base *CommandTransformBase) CommandWordsRequirer(plan *abs.Plan, arise *herr.AriseRef) func() herr.BuildError {
	return func() herr.BuildError {
		return base.RequireCommandWords(plan, arise)
	}
}

type SingleCommandTransform struct {
	con.SingleTransformBase
	CommandTransformBase
}

func (transform *SingleCommandTransform) Plan(destination abs.Artifact, plan *abs.Plan) herr.BuildError {
	return con.PlanSingleTransform(
		transform,
		transform.Source,
		destination,
		plan,
		transform.CommandWordsRequirer(plan, transform.Arise),
		func() herr.BuildError {
			transform.PlanCommandTransform(
				transform.Description,
				[]abs.Artifact{transform.Source},
				destination,
				plan,
				transform.TransformArise(),
			)
			return nil
		},
	)
}

func (transform *SingleCommandTransform) DumpTransform(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Print("exec ")
	con.PrintErrorString(prn, transform.Description)
	prn.Println(" {")
	prn.Indent(1)
	err := transform.DumpCommandLine(level + 1)
	if err != nil {
		prn.Fail(err)
	}
	prn.Println()
	if transform.Loud {
		prn.Indent(1)
		prn.Println("loud")
	}
	if transform.SuffixIsDestination {
		prn.Indent(1)
		prn.Println("suffixIsDestination")
	}
	prn.Indent(1)
	prn.Print("artifact ")
	con.PrintErrorString(prn, transform.Source.ArtifactKey().Unified())
	prn.Println()
	prn.Indent(0)
	prn.Print("}")
	return prn.Done()
}

var _ abs.Transform = &SingleCommandTransform{}

func NewSingleCommandTransform(
	description string,
	arise *herr.AriseRef,
	source abs.Artifact,
	commandLine VariableCommandLine,
	dumpCommandLine CommandLineDumper,
	requireCommandWords CommandWordsRequirer,
	loud bool,
	suffixIsDestination bool,
) *SingleCommandTransform {
	transform := &SingleCommandTransform {}
	transform.Description = description
	transform.Arise = arise
	transform.Source = source
	transform.CommandLine = commandLine
	transform.DumpCommandLine = dumpCommandLine
	transform.RequireCommandWords = requireCommandWords
	transform.Loud = loud
	transform.SuffixIsDestination = suffixIsDestination
	return transform
}

type MultiCommandTransform struct {
	con.MultiTransformBase
	CommandTransformBase
}

func (transform *MultiCommandTransform) Plan(destination abs.Artifact, plan *abs.Plan) herr.BuildError {
	return con.PlanMultiTransform(
		transform,
		transform.Sources,
		destination,
		plan,
		transform.CommandWordsRequirer(plan, transform.Arise),
		func() herr.BuildError {
			transform.PlanCommandTransform(
				transform.Description,
				transform.Sources,
				destination,
				plan,
				transform.TransformArise(),
			)
			return nil
		},
	)
}

func (transform *MultiCommandTransform) DumpTransform(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Level(level)
	prn.Print("exec ")
	con.PrintErrorString(prn, transform.Description)
	prn.Println(" {")
	prn.Indent(1)
	err := transform.DumpCommandLine(level + 1)
	if err != nil {
		prn.Fail(err)
	}
	prn.Println()
	if transform.Loud {
		prn.Indent(1)
		prn.Println("loud")
	}
	if transform.SuffixIsDestination {
		prn.Indent(1)
		prn.Println("suffixIsDestination")
	}
	for _, source := range transform.Sources {
		prn.Indent(1)
		prn.Print("artifact ")
		con.PrintErrorString(prn, source.ArtifactKey().Unified())
		prn.Println()
	}
	prn.Indent(0)
	prn.Print("}")
	return prn.Done()
}

var _ abs.Transform = &MultiCommandTransform{}

func NewMultiCommandTransform(
	description string,
	arise *herr.AriseRef,
	commandLine VariableCommandLine,
	dumpCommandLine CommandLineDumper,
	requireCommandWords CommandWordsRequirer,
	loud bool,
	suffixIsDestination bool,
) *MultiCommandTransform {
	transform := &MultiCommandTransform {}
	transform.Description = description
	transform.Arise = arise
	transform.CommandLine = commandLine
	transform.DumpCommandLine = dumpCommandLine
	transform.RequireCommandWords = requireCommandWords
	transform.Loud = loud
	transform.SuffixIsDestination = suffixIsDestination
	return transform
}

// ---------------------------------------- Action ----------------------------------------

type DeletePathAction struct {
	con.ActionBase
	Path string
	Base string
	Project string
}

func (action *DeletePathAction) SimpleDescr() string {
	return "delete " + con.GuessFileArtifactName(action.Path, action.Base)
}

func (action *DeletePathAction) Perform(plan *abs.Plan) herr.BuildError {
	step := &DeletePathStep {
		Path: action.Path,
		DeleteArise: action.Arise,
	}
	step.Description = "delete " + con.GuessFileArtifactName(action.Path, action.Base)
	plan.AddStep(step)
	return nil
}

var _ abs.Action = &DeletePathAction{}

type DeleteArtifactAction struct {
	con.ActionBase
	Artifact abs.Artifact
	Base string
}

func (action *DeleteArtifactAction) SimpleDescr() string {
	return fmt.Sprintf("[%s] delete %s", action.Artifact.ArtifactKey().Project, action.Artifact.DisplayName())
}

func (action *DeleteArtifactAction) Perform(plan *abs.Plan) herr.BuildError {
	step := &DeleteArtifactStep {
		Artifact: action.Artifact,
		DeleteArise: action.Arise,
	}
	step.Description = "delete " + action.Artifact.DisplayName()
	plan.AddStep(step)
	return nil
}

var _ abs.Action = &DeleteArtifactAction{}
