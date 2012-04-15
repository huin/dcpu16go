package core

import "fmt"

type ValueCodeError Word

func (err ValueCodeError) Error() string {
	return fmt.Sprintf("invalid value code 0x%02x", Word(err))
}

type Value interface {
	Write(MachineState, Word)
	Read(MachineState) Word
	LoadInstValue(WordLoader) error
	NumExtraWords() Word
	String() string
}

type ValueSet interface {
	Value(w Word) (Value, error)
}
