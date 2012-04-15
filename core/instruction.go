package model

import (
	"fmt"
)

type InvalidUnaryOpCodeError Word

func (err InvalidUnaryOpCodeError) Error() string {
	return fmt.Sprintf("invalid unary opcode 0x%02x", Word(err))
}

type InvalidBinaryOpCodeError Word

func (err InvalidBinaryOpCodeError) Error() string {
	return fmt.Sprintf("invalid binary opcode 0x%02x", Word(err))
}

type Instruction interface {
	LoadNextWords(WordLoader) error
	Execute(MachineState) error // TODO Return ticks.
	String() string
}

type InstructionSet interface {
	// Instruction returns the instruction for the given word. The returned
	// Instruction must have LoadNextWords called on it before it will function
	// properly.
	Instruction(word Word) (Instruction, error)

	// NumExtraWords returns the number of extra words required to be read for
	// the given instruction.
	NumExtraWords(word Word) (Word, error)
}

type UnaryInstruction interface {
	Instruction
	SetUnaryValue(Value)
}

type BinaryInstruction interface {
	Instruction
	SetBinaryValue(Value, Value)
}
