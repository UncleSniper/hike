package generic

import (
	"os"
	"strings"
	herr "hike/error"
	loc "hike/location"
	abs "hike/abstract"
	con "hike/concrete"
)

type CommandWord interface {
	AssembleCommand(
		sources []string,
		destinations []string,
		command [][]string,
		execArise *herr.AriseRef,
	) ([][]string, herr.BuildError)
	DumpCommandWord(level uint) error
	RequireCommandWordArtifacts(plan *abs.Plan, arise *herr.AriseRef) herr.BuildError
}

func mergeCommandLine(paths []string) (string, error) {
	var sink strings.Builder
	for index, path := range paths {
		if index > 0 {
			_, err := sink.WriteRune(os.PathListSeparator)
			if err != nil {
				return "", err
			}
		}
		_, err := sink.WriteString(path)
		if err != nil {
			return "", err
		}
	}
	return sink.String(), nil
}

func expandCommandLine(pieces []string, commands [][]string, merge bool) ([][]string, error) {
	if merge {
		merged, err := mergeCommandLine(pieces)
		if err != nil {
			return nil, err
		}
		pieces = []string {
			merged,
		}
	}
	var result [][]string
	var line []string
	for _, command := range commands {
		for _, piece := range pieces {
			line = nil
			for _, w := range command {
				line = append(line, w)
			}
			result = append(result, append(line, piece))
		}
	}
	return result, nil
}

func joinCommandLine(commands [][]string, line []string) []string {
	for _, command := range commands {
		for _, word := range command {
			line = append(line, word)
		}
	}
	return line
}

func AssembleCommand(
	sources []string,
	destinations []string,
	words []CommandWord,
	execArise *herr.AriseRef,
) (line []string, err herr.BuildError) {
	var ch [][]string
	for _, word := range words {
		ch, err = word.AssembleCommand(sources, destinations, [][]string{nil}, execArise)
		if err != nil {
			return
		}
		line = joinCommandLine(ch, line)
	}
	return
}

type CommandWordBase struct {
	Merge bool
}

// ---------------------------------------- AssembleCommandError ----------------------------------------

type AssembleCommandError struct {
	herr.BuildErrorBase
	Fault error
	ExecArise *herr.AriseRef
}

func (failed *AssembleCommandError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Printf("Failed to assemble command: %s\n", failed.Fault.Error())
	prn.Indent(0)
	prn.Print("for execution ")
	prn.Arise(failed.ExecArise, 0)
	failed.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (failed *AssembleCommandError) BuildErrorLocation() *loc.Location {
	return failed.ExecArise.Location
}

var _ herr.BuildError = &AssembleCommandError{}

// ---------------------------------------- StaticCommandWord ----------------------------------------

type StaticCommandWord struct {
	Word string
}

func (word *StaticCommandWord) AssembleCommand(
	sources []string,
	destinations []string,
	commands [][]string,
	execArise *herr.AriseRef,
) ([][]string, herr.BuildError) {
	for index, command := range commands {
		commands[index] = append(command, word.Word)
	}
	return commands, nil
}

func (word *StaticCommandWord) DumpCommandWord(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	con.PrintErrorString(prn, word.Word)
	return prn.Done()
}

func (word *StaticCommandWord) RequireCommandWordArtifacts(
	plan *abs.Plan,
	arise *herr.AriseRef,
) herr.BuildError {
	return nil
}

var _ CommandWord = &StaticCommandWord{}

// ---------------------------------------- SourceCommandWord ----------------------------------------

type SourceCommandWord struct {
	CommandWordBase
}

func (word *SourceCommandWord) AssembleCommand(
	sources []string,
	destinations []string,
	commands [][]string,
	execArise *herr.AriseRef,
) ([][]string, herr.BuildError) {
	result, err := expandCommandLine(sources, commands, word.Merge)
	if err != nil {
		return nil, &AssembleCommandError {
			Fault: err,
			ExecArise: execArise,
		}
	}
	return result, nil
}

func (word *SourceCommandWord) DumpCommandWord(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Print("source")
	if word.Merge {
		prn.Print(" merge")
	}
	return prn.Done()
}

func (word *SourceCommandWord) RequireCommandWordArtifacts(
	plan *abs.Plan,
	arise *herr.AriseRef,
) herr.BuildError {
	return nil
}

var _ CommandWord = &SourceCommandWord{}

// ---------------------------------------- DestinationCommandWord ----------------------------------------

type DestinationCommandWord struct {
	CommandWordBase
}

func (word *DestinationCommandWord) AssembleCommand(
	sources []string,
	destinations []string,
	commands [][]string,
	execArise *herr.AriseRef,
) ([][]string, herr.BuildError) {
	result, err := expandCommandLine(destinations, commands, word.Merge)
	if err != nil {
		return nil, &AssembleCommandError {
			Fault: err,
			ExecArise: execArise,
		}
	}
	return result, nil
}

func (word *DestinationCommandWord) DumpCommandWord(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Print("dest")
	if word.Merge {
		prn.Print(" merge")
	}
	return prn.Done()
}

func (word *DestinationCommandWord) RequireCommandWordArtifacts(
	plan *abs.Plan,
	arise *herr.AriseRef,
) herr.BuildError {
	return nil
}

var _ CommandWord = &DestinationCommandWord{}

// ---------------------------------------- ArtifactCommandWord ----------------------------------------

type ArtifactCommandWord struct {
	CommandWordBase
	Artifact abs.Artifact
}

func (word *ArtifactCommandWord) AssembleCommand(
	sources []string,
	destinations []string,
	commands [][]string,
	execArise *herr.AriseRef,
) ([][]string, herr.BuildError) {
	paths, err := word.Artifact.PathNames(nil)
	if err != nil {
		return nil, err
	}
	result, eerr := expandCommandLine(paths, commands, word.Merge)
	if eerr != nil {
		return nil, &AssembleCommandError {
			Fault: eerr,
			ExecArise: execArise,
		}
	}
	return result, nil
}

func (word *ArtifactCommandWord) DumpCommandWord(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Print("aux ")
	con.PrintErrorString(prn, word.Artifact.ArtifactKey().Unified())
	if word.Merge {
		prn.Print(" merge")
	}
	return prn.Done()
}

func (word *ArtifactCommandWord) RequireCommandWordArtifacts(
	plan *abs.Plan,
	arise *herr.AriseRef,
) herr.BuildError {
	return word.Artifact.Require(plan, arise)
}

var _ CommandWord = &ArtifactCommandWord{}

// ---------------------------------------- BraceCommandWord ----------------------------------------

type BraceCommandWord struct {
	Children []CommandWord
}

func (word *BraceCommandWord) AddChild(child CommandWord) {
	word.Children = append(word.Children, child)
}

func (word *BraceCommandWord) AssembleCommand(
	sources []string,
	destinations []string,
	commands [][]string,
	execArise *herr.AriseRef,
) (result [][]string, err herr.BuildError) {
	nested := [][]string{nil}
	for _, child := range word.Children {
		nested, err = child.AssembleCommand(sources, destinations, nested, execArise)
		if err != nil {
			return
		}
	}
	content := joinCommandLine(nested, nil)
	for index, command := range commands {
		for _, w := range content {
			command = append(command, w)
		}
		commands[index] = command
	}
	result = commands
	return
}

func (word *BraceCommandWord) DumpCommandWord(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Print("{")
	for index, child := range word.Children {
		if index > 0 {
			prn.Print(" ")
		}
		err := child.DumpCommandWord(level)
		if err != nil {
			prn.Fail(err)
		}
	}
	prn.Print("}")
	return prn.Done()
}

func (word *BraceCommandWord) RequireCommandWordArtifacts(
	plan *abs.Plan,
	arise *herr.AriseRef,
) herr.BuildError {
	for _, child := range word.Children {
		err := child.RequireCommandWordArtifacts(plan, arise)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ CommandWord = &BraceCommandWord{}

// ---------------------------------------- misc ----------------------------------------

func DumpCommandWords(words []CommandWord, level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	for index, word := range words {
		if index > 0 {
			prn.Print(" ")
		}
		err := word.DumpCommandWord(level)
		if err != nil {
			prn.Fail(err)
		}
	}
	return prn.Done()
}
