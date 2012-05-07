package core

import (
	"errors"
	"fmt"
)

type ValueCodeError Word

func (err ValueCodeError) Error() string {
	return fmt.Sprintf("invalid value code 0x%02x", Word(err))
}

var ValueContextError = errors.New("value cannot be used in that context")

type Value interface {
	Write(MachineState, Word)
	Read(MachineState) Word
	LoadExtraWords(WordLoader) error
	NumExtraWords() Word
	// Create a copy of the value.
	Clone() Value
	String() string
}

type ValueLiteral interface {
	Value
	SetLiteral(Word)
}

type ValueSet interface {
	Value(w Word, asValueB bool) (Value, error)
}
