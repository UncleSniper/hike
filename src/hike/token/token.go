package token

import "strings"
import "strconv"
import loc "hike/location"

type Type uint

const (
	T_NAME Type = iota
	T_STRING
	T_INT
	T_LBRACE
	T_RBRACE
	T_EOF
)

type Token struct {
	Location loc.Location
	Type Type
	Text string
}

func (token *Token) ReconstructTo(sink *strings.Builder) error {
	var err error
	switch token.Type {
		case T_STRING:
			_, err = sink.WriteString(strconv.Quote(token.Text))
			return err
		case T_EOF:
			_, err = sink.WriteString("<end of input>")
			return err
		default:
			_, err = sink.WriteRune('\'')
			if err != nil {
				return err
			}
			_, err = sink.WriteString(token.Text)
			if err != nil {
				return err
			}
			_, err = sink.WriteRune('\'')
			return err
	}
}

func (token *Token) Reconstruct() (string, error) {
	var sink strings.Builder
	err := token.ReconstructTo(&sink)
	if err == nil {
		return sink.String(), nil
	} else {
		return "", err
	}
}
