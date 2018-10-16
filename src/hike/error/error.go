package error

import (
	"os"
	"io"
	"fmt"
	loc "hike/location"
)

type AriseRef struct {
	Text string
	Location *loc.Location
}

func (ref *AriseRef) PrintArise(level uint) error {
	prn := &ErrorPrinter{}
	prn.Println("arising from")
	prn.Indent(level + 1)
	prn.Println(ref.Text)
	prn.Indent(level)
	location, err := ref.Location.Format()
	if err != nil {
		return err
	}
	prn.Printf("at %s", location)
	return prn.Done()
}

func IndentError(level uint, out io.Writer) (err error) {
	for ; level > 0; level-- {
		_, err = fmt.Fprint(out, "    ")
		if err != nil {
			break
		}
	}
	return
}

type BuildFrame interface {
	PrintErrorFrame(level uint) error
}

type BuildError interface {
	PrintBuildError(level uint) error
	AddErrorFrame(frame BuildFrame)
	BuildErrorLocation() *loc.Location
}

type BuildErrorBase struct {
	frames []BuildFrame
}

func (err *BuildErrorBase) AddErrorFrame(frame BuildFrame) {
	err.frames = append(err.frames, frame)
}

func (base *BuildErrorBase) PrintBacktrace(level uint) error {
	count := len(base.frames)
	prn := &ErrorPrinter{}
	prn.Level(level)
	for i := 0; i < count; i++ {
		prn.Println()
		prn.Indent(0)
		prn.Frame(base.frames[i], 1)
	}
	return prn.Done()
}

func (base *BuildErrorBase) InjectBacktrace(printer *ErrorPrinter, level uint) {
	printer.Inject(func(innerLevel uint) error {
		return base.PrintBacktrace(innerLevel)
	}, level)
}

func NewErrorPrinter() *ErrorPrinter {
	return &ErrorPrinter {
		Out: os.Stderr,
	}
}

type ErrorPrinter struct {
	firstError error
	level uint
	Out io.Writer
}

func (printer *ErrorPrinter) Print(values ...interface{}) {
	if printer.firstError == nil {
		_, printer.firstError = fmt.Fprint(printer.Out, values...)
	}
}

func (printer *ErrorPrinter) Println(values ...interface{}) {
	if printer.firstError == nil {
		_, printer.firstError = fmt.Fprintln(printer.Out, values...)
	}
}

func (printer *ErrorPrinter) Printf(format string, values ...interface{}) {
	if printer.firstError == nil {
		_, printer.firstError = fmt.Fprintf(printer.Out, format, values...)
	}
}

func (printer *ErrorPrinter) Level(level uint) {
	printer.level = level
}

func (printer *ErrorPrinter) Indent(level uint) {
	if printer.firstError == nil {
		printer.firstError = IndentError(printer.level + level, printer.Out)
	}
}

func (printer *ErrorPrinter) Arise(arise *AriseRef, level uint) {
	if printer.firstError == nil {
		printer.firstError = arise.PrintArise(printer.level + level)
	}
}

func (printer *ErrorPrinter) Frame(frame BuildFrame, level uint) {
	if printer.firstError == nil {
		printer.firstError = frame.PrintErrorFrame(printer.level + level)
	}
}

func (printer *ErrorPrinter) Location(location *loc.Location) {
	if printer.firstError == nil {
		str, err := location.Format()
		if err != nil {
			printer.firstError = err
			return
		}
		_, printer.firstError = fmt.Fprint(printer.Out, str)
	}
}

func (printer *ErrorPrinter) Inject(callback func(level uint) error, level uint) {
	if printer.firstError == nil {
		err := callback(printer.level + level)
		if printer.firstError == nil {
			printer.firstError = err
		}
	}
}

func (printer *ErrorPrinter) Done() error {
	return printer.firstError
}
