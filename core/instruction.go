package core

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
	// Create a copy of the instruction.
	Clone() Instruction
	String() string
}

type UnaryInstruction interface {
	Instruction
	SetUnaryValue(Value)
}

type BinaryInstruction interface {
	Instruction
	SetBinaryValue(Value, Value)
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

func InstructionSkip(wordLoader WordLoader, set InstructionSet) error {
	word, err := wordLoader.WordLoad()
	if err != nil {
		return err
	}
	count, err := set.NumExtraWords(word)
	if err != nil {
		return err
	}
	return wordLoader.SkipWords(count)
}

func InstructionLoad(wordLoader WordLoader, set InstructionSet) (Instruction, error) {
	word, err := wordLoader.WordLoad()
	if err != nil {
		return nil, err
	}
	instruction, err := set.Instruction(word)
	if err != nil {
		return nil, err
	}
	err = instruction.LoadNextWords(wordLoader)
	if err != nil {
		return nil, err
	}
	return instruction, err
}
