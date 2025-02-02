package concatenation

import (
	ebnf "github.com/almerlucke/exbana/v2"
	"io"
)

// Concatenation matches a series of patterns AND style in order (concatenation)
type Concatenation[T, P any] struct {
	*ebnf.BasePattern[T, P]
	patterns ebnf.Patterns[T, P]
}

// New creates a new concatenation pattern
func New[T, P any](patterns ...ebnf.Pattern[T, P]) *Concatenation[T, P] {
	c := &Concatenation[T, P]{
		BasePattern: ebnf.NewBasePattern[T, P](),
		patterns:    patterns,
	}

	c.SetSelf(c)

	return c
}

// Match matches AND against a stream, fails if any of the sub patterns mismatches
func (c *Concatenation[T, P]) Match(rd ebnf.Reader[T, P]) (bool, *ebnf.Match[T, P], error) {
	var matches []*ebnf.Match[T, P]

	beginPos, err := rd.Position()
	if ebnf.IsStreamError(err) {
		return false, nil, err
	}

	for _, pm := range c.patterns {
		subBeginPos, err := rd.Position()
		if ebnf.IsStreamError(err) {
			return false, nil, err
		}

		matched, result, err := pm.Match(rd)
		if err != nil {
			return false, nil, err
		}

		if matched {
			matches = append(matches, result)
		} else {
			subEndPos, err := rd.Position()
			if ebnf.IsStreamError(err) {
				return false, nil, err
			}

			c.Logger().LogMismatch(ebnf.NewMismatch(c, beginPos, subEndPos, ebnf.NewMatch(pm, subBeginPos, subEndPos, nil, nil), matches))

			return false, nil, nil
		}
	}

	endPos, err := rd.Position()
	if ebnf.IsStreamError(err) {
		return false, nil, err
	}

	return true, ebnf.NewMatch(c, beginPos, endPos, nil, matches), nil
}

// Generate writes a concatenation of patterns to a writer
func (c *Concatenation[T, P]) Generate(w ebnf.Writer[T]) error {
	for _, child := range c.patterns {
		err := child.Generate(w)
		if err != nil {
			return err
		}
	}

	return nil
}

// Print EBNF concatenation group
func (c *Concatenation[T, P]) Print(w io.Writer) error {
	_, err := w.Write([]byte("("))
	if err != nil {
		return err
	}

	first := true

	for _, child := range c.patterns {
		if !first {
			_, err = w.Write([]byte(", "))
			if err != nil {
				return err
			}
		}

		err = child.PrintAsChild(w)
		if err != nil {
			return err
		}

		first = false
	}

	_, err = w.Write([]byte(")"))

	return err
}
