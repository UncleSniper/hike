package generic

import (
	"os"
	herr "hike/error"
	con "hike/concrete"
)

type CommandWord interface {
	AssembleCommand(sources []string, destinations []string, command [][]string) [][]string
	DumpCommandWord(level uint) error
}

func expandCommandLine(pieces []string, commands [][]string) [][]string {
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
	return result
}

func joinCommandLine(commands [][]string, line []string) []string {
	for _, command := range commands {
		for _, word := range command {
			line = append(line, word)
		}
	}
	return line
}

func AssembleCommand(sources []string, destinations []string, words []CommandWord) []string {
	var line []string
	for _, word := range words {
		line = joinCommandLine(word.AssembleCommand(sources, destinations, [][]string{nil}), line)
	}
	return line
}

// ---------------------------------------- StaticCommandWord ----------------------------------------

type StaticCommandWord struct {
	Word string
}

func (word *StaticCommandWord) AssembleCommand(
	sources []string,
	destinations []string,
	commands [][]string,
) [][]string {
	for index, command := range commands {
		commands[index] = append(command, word.Word)
	}
	return commands
}

func (word *StaticCommandWord) DumpCommandWord(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	con.PrintErrorString(prn, word.Word)
	return prn.Done()
}

var _ CommandWord = &StaticCommandWord{}

// ---------------------------------------- SourceCommandWord ----------------------------------------

type SourceCommandWord struct {}

func (word *SourceCommandWord) AssembleCommand (
	sources []string,
	destinations []string,
	commands [][]string,
) [][]string {
	return expandCommandLine(sources, commands)
}

func (word *SourceCommandWord) DumpCommandWord(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Print("source")
	return prn.Done()
}

var _ CommandWord = &SourceCommandWord{}

// ---------------------------------------- DestinationCommandWord ----------------------------------------

type DestinationCommandWord struct {}

func (word *DestinationCommandWord) AssembleCommand (
	sources []string,
	destinations []string,
	commands [][]string,
) [][]string {
	return expandCommandLine(destinations, commands)
}

func (word *DestinationCommandWord) DumpCommandWord(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Out = os.Stdout
	prn.Print("dest")
	return prn.Done()
}

var _ CommandWord = &DestinationCommandWord{}

// ---------------------------------------- BraceCommandWord ----------------------------------------

type BraceCommandWord struct {
	Children []CommandWord
}

func (word *BraceCommandWord) AddChild(child CommandWord) {
	word.Children = append(word.Children, child)
}

func (word *BraceCommandWord) AssembleCommand (
	sources []string,
	destinations []string,
	commands [][]string,
) [][]string {
	nested := [][]string{nil}
	for _, child := range word.Children {
		nested = child.AssembleCommand(sources, destinations, nested)
	}
	content := joinCommandLine(nested, nil)
	for index, command := range commands {
		for _, w := range content {
			command = append(command, w)
		}
		commands[index] = command
	}
	return commands
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

var _ CommandWord = &BraceCommandWord{}

// ---------------------------------------- BraceCommandWord ----------------------------------------

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
