package location

import "fmt"
import "strings"

type Location struct {
	File string
	Line uint
	Column uint
}

func (location *Location) FormatTo(sink *strings.Builder) error {
	if len([]rune(location.File)) == 0 {
		_, err := sink.WriteString("<unknown location>")
		return err
	}
	_, err := sink.WriteString(location.File)
	if err != nil {
		return err
	}
	if location.Line > 0 {
		_, err = sink.WriteRune(':')
		if err != nil {
			return err
		}
		_, err = sink.WriteString(fmt.Sprint(location.Line))
		if err != nil {
			return err
		}
		if location.Column > 0 {
			_, err = sink.WriteRune(':')
			if err != nil {
				return err
			}
			_, err = sink.WriteString(fmt.Sprint(location.Column))
			return err
		}
		return nil
	} else {
		_, err = sink.WriteString(":<unkown line>")
		return err
	}
}

func (location *Location) Format() (string, error) {
	var sink strings.Builder
	err := location.FormatTo(&sink)
	if err == nil {
		return sink.String(), nil
	} else {
		return "", err
	}
}
