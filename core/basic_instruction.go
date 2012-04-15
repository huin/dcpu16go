package model

import (
	"fmt"
)

func splitOpWord(w Word) (upper6, middle6, lower4 Word) {
	upper6 = (w >> 10) & 0x003f
	middle6 = (w >> 4) & 0x003f
	lower4 = w & 0x000f
	return
}

// BasicInstructionSet is the table of basic instructions supported by the
// DCPU-16. This implementation avoids memory allocation, but the instruction
// returned is only good for use until the next call to the Instruction method.
type BasicInstructionSet struct {
	initialized bool
	jsrInst     JsrInst
	binarySet   [0x10]BinaryInstruction
	setInst     SetInst
	addInst     AddInst
	subInst     SubInst
	mulInst     MulInst
	divInst     DivInst
	modInst     ModInst
	shlInst     ShlInst
	shrInst     ShrInst
	andInst     AndInst
	borInst     BorInst
	xorInst     XorInst
	ifeInst     IfeInst
	ifnInst     IfnInst
	ifgInst     IfgInst
	ifbInst     IfbInst

	// Use separate pools of values so that two values in play at the same time
	// don't interfere (see doc for BasicValueSet).
	aValueSet BasicValueSet
	bValueSet BasicValueSet
}

func (is *BasicInstructionSet) init() {
	is.binarySet = [0x10]BinaryInstruction{
		nil, // Indicates unary instruction.
		&is.setInst, &is.addInst, &is.subInst, &is.mulInst,
		&is.divInst, &is.modInst, &is.shlInst, &is.shrInst,
		&is.andInst, &is.borInst, &is.xorInst, &is.ifeInst,
		&is.ifnInst, &is.ifgInst, &is.ifbInst,
	}
}

func (is *BasicInstructionSet) unaryInstruction(upper6, middle6 Word) (instruction UnaryInstruction, a Value, err error) {
	if middle6 != 0x001 {
		err = InvalidUnaryOpCodeError(middle6)
		return
	}
	instruction = &is.jsrInst
	a, err = is.aValueSet.Value(upper6)
	return
}

func (is *BasicInstructionSet) binaryInstruction(upper6, middle6, lower4 Word) (instruction BinaryInstruction, a, b Value, err error) {
	opCode := lower4
	if int(opCode) >= len(is.binarySet) {
		// Shouldn't happen with the given set.
		err = InvalidBinaryOpCodeError(opCode)
		return
	}
	instruction = is.binarySet[opCode]
	if instruction == nil {
		// Shouldn't happen with the given set.
		err = InvalidBinaryOpCodeError(opCode)
		return
	}
	a, err = is.aValueSet.Value(middle6)
	if err != nil {
		return
	}
	b, err = is.bValueSet.Value(upper6)
	if err != nil {
		return
	}
	return
}

func (is *BasicInstructionSet) NumExtraWords(w Word) (Word, error) {
	if !is.initialized {
		is.init()
	}

	upper6, middle6, lower4 := splitOpWord(w)

	if lower4 != 0 {
		// Binary instruction.
		_, a, b, err := is.binaryInstruction(upper6, middle6, lower4)
		if err != nil {
			return 0, err
		}
		return a.NumExtraWords() + b.NumExtraWords(), nil
	}

	// Unary instruction.
	_, a, err := is.unaryInstruction(upper6, middle6)
	if err != nil {
		return 0, err
	}
	return a.NumExtraWords(), nil
}

func (is *BasicInstructionSet) Instruction(w Word) (Instruction, error) {
	if !is.initialized {
		is.init()
	}

	upper6, middle6, lower4 := splitOpWord(w)

	if lower4 != 0 {
		// Binary instruction.
		instruction, a, b, err := is.binaryInstruction(upper6, middle6, lower4)
		if err != nil {
			return nil, err
		}
		instruction.SetBinaryValue(a, b)
		return instruction, nil
	}

	// Unary instruction.
	instruction, a, err := is.unaryInstruction(upper6, middle6)
	if err != nil {
		return nil, err
	}
	instruction.SetUnaryValue(a)
	return instruction, nil
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

// unaryInst forms common data and code for instructions that take one value.
type unaryInst struct {
	A Value
}

func (o *unaryInst) SetUnaryValue(a Value) {
	o.A = a
}

func (o *unaryInst) LoadNextWords(wordLoader WordLoader) error {
	return o.A.LoadInstValue(wordLoader)
}

func (o *unaryInst) format(name string) string {
	return fmt.Sprintf("%s %v", name, o.A)
}

// 0x01: JSR a - pushes the address of the next instruction to the stack, then sets PC to a
type JsrInst struct {
	unaryInst
}

func (o *JsrInst) Execute(state MachineState) error {
	state.WriteMemory(state.DecReadSP(), state.PC())
	state.WritePC(o.A.Read(state))
	return nil
}

func (o *JsrInst) String() string {
	return o.unaryInst.format("JSR")
}

// binaryInst forms common data and code for instructions that take two values (A
// and B).
type binaryInst struct {
	A, B Value
}

func (o *binaryInst) SetBinaryValue(a, b Value) {
	o.A = a
	o.B = b
}

func (o *binaryInst) LoadNextWords(wordLoader WordLoader) error {
	err := o.A.LoadInstValue(wordLoader)
	if err != nil {
		return err
	}
	return o.B.LoadInstValue(wordLoader)
}

func (o *binaryInst) format(name string) string {
	return fmt.Sprintf("%s %v, %v", name, o.A, o.B)
}

// 0x1: SET a, b - sets a to b
type SetInst struct {
	binaryInst
}

func (o *SetInst) Execute(state MachineState) error {
	o.A.Write(state, o.B.Read(state))
	return nil
}

func (o *SetInst) String() string {
	return o.binaryInst.format("SET")
}

// 0x2: ADD a, b - sets a to a+b, sets O to 0x0001 if there's an overflow, 0x0 otherwise
type AddInst struct {
	binaryInst
}

func (o *AddInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result := uint32(a) + uint32(b)
	o.A.Write(state, Word(result&0xffff))
	state.WriteO(Word(result >> 16))
	return nil
}

func (o *AddInst) String() string {
	return o.binaryInst.format("ADD")
}

// 0x3: SUB a, b - sets a to a-b, sets O to 0xffff if there's an underflow, 0x0 otherwise
type SubInst struct {
	binaryInst
}

func (o *SubInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result := uint32(a) - uint32(b)
	o.A.Write(state, Word(result&0xffff))
	state.WriteO(Word(result >> 16))
	return nil
}

func (o *SubInst) String() string {
	return o.binaryInst.format("SUB")
}

// 0x4: MUL a, b - sets a to a*b, sets O to ((a*b)>>16)&0xffff
type MulInst struct {
	binaryInst
}

func (o *MulInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result := uint32(a) * uint32(b)
	o.A.Write(state, Word(result&0xffff))
	state.WriteO(Word(result >> 16))
	return nil
}

func (o *MulInst) String() string {
	return o.binaryInst.format("MUL")
}

// 0x5: DIV a, b - sets a to a/b, sets O to ((a<<16)/b)&0xffff. if b==0, sets a and O to 0 instructionead.
type DivInst struct {
	binaryInst
}

func (o *DivInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result := (uint32(a) << 16) / uint32(b)
	o.A.Write(state, Word(result>>16))
	state.WriteO(Word(result & 0xffff))
	return nil
}

func (o *DivInst) String() string {
	return o.binaryInst.format("DIV")
}

// 0x6: MOD a, b - sets a to a%b. if b==0, sets a to 0 instructionead.
type ModInst struct {
	binaryInst
}

func (o *ModInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	o.A.Write(state, a%b)
	return nil
}

func (o *ModInst) String() string {
	return o.binaryInst.format("MOD")
}

// 0x7: SHL a, b - sets a to a<<b, sets O to ((a<<b)>>16)&0xffff
type ShlInst struct {
	binaryInst
}

func (o *ShlInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result := uint32(a) << uint32(b)
	o.A.Write(state, Word(result))
	state.WriteO(Word(result >> 16))
	return nil
}

func (o *ShlInst) String() string {
	return o.binaryInst.format("SHL")
}

// 0x8: SHR a, b - sets a to a>>b, sets O to ((a<<16)>>b)&0xffff
type ShrInst struct {
	binaryInst
}

func (o *ShrInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result := (uint32(a) << 16) >> uint32(b)
	o.A.Write(state, Word(result>>16))
	state.WriteO(Word(result & 0xffff))
	return nil
}

func (o *ShrInst) String() string {
	return o.binaryInst.format("SHR")
}

// 0x9: AND a, b - sets a to a&b
type AndInst struct {
	binaryInst
}

func (o *AndInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	o.A.Write(state, a&b)
	return nil
}

func (o *AndInst) String() string {
	return o.binaryInst.format("AND")
}

// 0xa: BOR a, b - sets a to a|b
type BorInst struct {
	binaryInst
}

func (o *BorInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	o.A.Write(state, a|b)
	return nil
}

func (o *BorInst) String() string {
	return o.binaryInst.format("BOR")
}

// 0xb: XOR a, b - sets a to a^b
type XorInst struct {
	binaryInst
}

func (o *XorInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	o.A.Write(state, a^b)
	return nil
}

func (o *XorInst) String() string {
	return o.binaryInst.format("XOR")
}

// 0xc: IFE a, b - performs next instruction only if a==b
type IfeInst struct {
	binaryInst
}

func (o *IfeInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if a == b {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfeInst) String() string {
	return o.binaryInst.format("IFE")
}

// 0xd: IFN a, b - performs next instruction only if a!=b
type IfnInst struct {
	binaryInst
}

func (o *IfnInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if a != b {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfnInst) String() string {
	return o.binaryInst.format("IFN")
}

// 0xe: IFG a, b - performs next instruction only if a>b
type IfgInst struct {
	binaryInst
}

func (o *IfgInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if a > b {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfgInst) String() string {
	return o.binaryInst.format("IFG")
}

// 0xf: IFB a, b - performs next instruction only if (a&b)!=0
type IfbInst struct {
	binaryInst
}

func (o *IfbInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if (a & b) != 0 {
		return InstructionSkip(state, state)
	}
	return nil
}

func (o *IfbInst) String() string {
	return o.binaryInst.format("IFB")
}
