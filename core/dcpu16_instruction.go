package core

import (
	"fmt"
)

func splitOpWord(w Word) (upper6, middle5, lower5 Word) {
	upper6 = (w >> 10) & 0x003f
	middle5 = (w >> 5) & 0x001f
	lower5 = w & 0x001f
	return
}

// D16InstructionSet is the table of basic instructions supported by the
// DCPU-16.
type D16InstructionSet struct {
	initialized bool
	jsrInst     JsrInst
	// TODO 1.7 unary instructions.

	binarySet [0x20]BinaryInstruction

	setInst SetInst

	addInst AddInst
	subInst SubInst
	mulInst MulInst
	mliInst MliInst
	divInst DivInst
	dviInst DviInst
	modInst ModInst
	mdiInst MdiInst

	andInst AndInst
	borInst BorInst
	xorInst XorInst
	shrInst ShrInst
	asrInst AsrInst
	shlInst ShlInst

	ifbInst IfbInst
	ifcInst IfcInst
	ifeInst IfeInst
	ifnInst IfnInst
	ifgInst IfgInst
	ifaInst IfaInst
	iflInst IflInst
	ifuInst IfuInst

	//adxInst AdxInst  // TODO
	//sbxInst SbxInst  // TODO

	//stiInst StiInst  // TODO
	//stdInst StdInst  // TODO

	// Use separate pools of values so that two values in play at the same time
	// don't interfere (see doc for D16ValueSet).
	aValueSet D16ValueSet
	bValueSet D16ValueSet
}

func (is *D16InstructionSet) init() {
	is.binarySet = [0x20]BinaryInstruction{
		// 0x00
		nil,

		// 0x01
		&is.setInst,

		// 0x02+
		&is.addInst, &is.subInst,
		&is.mulInst, &is.mliInst,
		&is.divInst, &is.dviInst,
		&is.modInst, &is.mdiInst,

		// 0x0a+
		&is.andInst, &is.borInst, &is.xorInst,
		&is.shrInst, &is.asrInst, &is.shlInst,

		// 0x10+
		&is.ifbInst, &is.ifcInst,
		&is.ifeInst, &is.ifnInst,
		&is.ifgInst, &is.ifaInst,
		&is.iflInst, &is.ifuInst,

		// 0x18+
		nil, nil,

		// 0x1a+
		nil /*TODO adxInst*/, nil, /*TODO sbxInst*/

		// 0x1c+
		nil, nil,

		// 0x1e+
		nil /*TODO stiInst*/, nil, /*TODO stdInst*/
	}
}

func (is *D16InstructionSet) unaryInstruction(upper6, middle5 Word) (instruction UnaryInstruction, a Value, err error) {
	if middle5 != 0x001 {
		err = InvalidUnaryOpCodeError(middle5)
		return
	}
	instruction = &is.jsrInst
	a, err = is.aValueSet.Value(upper6, false)
	return
}

func (is *D16InstructionSet) binaryInstruction(upper6, middle5, lower5 Word) (instruction BinaryInstruction, a, b Value, err error) {
	opCode := lower5
	if int(opCode) >= len(is.binarySet) {
		err = InvalidBinaryOpCodeError(opCode)
		return
	}
	instruction = is.binarySet[opCode]
	if instruction == nil {
		err = InvalidBinaryOpCodeError(opCode)
		return
	}
	a, err = is.aValueSet.Value(upper6, false)
	if err != nil {
		return
	}
	b, err = is.bValueSet.Value(middle5, true)
	if err != nil {
		return
	}
	return
}

func (is *D16InstructionSet) NumExtraWords(w Word) (Word, error) {
	if !is.initialized {
		is.init()
	}

	upper6, middle5, lower5 := splitOpWord(w)

	if lower5 != 0 {
		// Binary instruction.
		_, a, b, err := is.binaryInstruction(upper6, middle5, lower5)
		if err != nil {
			return 0, err
		}
		return a.NumExtraWords() + b.NumExtraWords(), nil
	}

	// Unary instruction.
	_, a, err := is.unaryInstruction(upper6, middle5)
	if err != nil {
		return 0, err
	}
	return a.NumExtraWords(), nil
}

func (is *D16InstructionSet) Instruction(w Word) (Instruction, error) {
	if !is.initialized {
		is.init()
	}

	upper6, middle5, lower5 := splitOpWord(w)

	if lower5 != 0 {
		// Binary instruction.
		instruction, a, b, err := is.binaryInstruction(upper6, middle5, lower5)
		if err != nil {
			return nil, err
		}
		instruction.SetBinaryValue(a, b)
		return instruction, nil
	}

	// Unary instruction.
	instruction, a, err := is.unaryInstruction(upper6, middle5)
	if err != nil {
		return nil, err
	}
	instruction.SetUnaryValue(a)
	return instruction, nil
}

func (is *D16InstructionSet) InstructionByName(name string) (Instruction, bool) {
	panic("unimplemented")
}

// unaryInst forms common data and code for instructions that take one value.
type unaryInst struct {
	A Value
}

func (o *unaryInst) SetUnaryValue(a Value) {
	o.A = a
}

func (o *unaryInst) LoadNextWords(wordLoader WordLoader) error {
	return o.A.LoadExtraWords(wordLoader)
}

func (o *unaryInst) clone() unaryInst {
	return unaryInst{A: o.A.Clone()}
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

func (o *JsrInst) Clone() Instruction {
	return &JsrInst{o.unaryInst.clone()}
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
	err := o.A.LoadExtraWords(wordLoader)
	if err != nil {
		return err
	}
	return o.B.LoadExtraWords(wordLoader)
}

func (o *binaryInst) clone() binaryInst {
	return binaryInst{A: o.A.Clone(), B: o.B.Clone()}
}

func (o *binaryInst) format(name string) string {
	return fmt.Sprintf("%s %v, %v", name, o.B, o.A)
}

// 0x01: SET b, a - sets b to a
type SetInst struct {
	binaryInst
}

func (o *SetInst) Execute(state MachineState) error {
	o.B.Write(state, o.A.Read(state))
	return nil
}

func (o *SetInst) Clone() Instruction {
	return &SetInst{o.binaryInst.clone()}
}

func (o *SetInst) String() string {
	return o.binaryInst.format("SET")
}

// 0x02: ADD b, a - sets b to b+a, sets EX to 0x0001 if there's an overflow, 0x0 otherwise
type AddInst struct {
	binaryInst
}

func (o *AddInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	ex, result := (DWord(b) + DWord(a)).Split()
	o.B.Write(state, result)
	state.WriteEX(ex)
	return nil
}

func (o *AddInst) Clone() Instruction {
	return &AddInst{o.binaryInst.clone()}
}

func (o *AddInst) String() string {
	return o.binaryInst.format("ADD")
}

// 0x03: SUB b, a - sets b to b-a, sets EX to 0xffff if there's an underflow, 0x0 otherwise
type SubInst struct {
	binaryInst
}

func (o *SubInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	ex, result := (DWord(b) - DWord(a)).Split()
	o.B.Write(state, result)
	state.WriteEX(ex)
	return nil
}

func (o *SubInst) Clone() Instruction {
	return &SubInst{o.binaryInst.clone()}
}

func (o *SubInst) String() string {
	return o.binaryInst.format("SUB")
}

// 0x04: MUL b, a - sets b to b*a, sets EX to ((b*a)>>16)&0xffff (treats b, a as
// unsigned)
type MulInst struct {
	binaryInst
}

func (o *MulInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	ex, result := (DWord(b) * DWord(a)).Split()
	o.B.Write(state, result)
	state.WriteEX(ex)
	return nil
}

func (o *MulInst) Clone() Instruction {
	return &MulInst{o.binaryInst.clone()}
}

func (o *MulInst) String() string {
	return o.binaryInst.format("MUL")
}

// 0x05: MLI b, a - like MUL, but treat b, a as signed
type MliInst struct {
	binaryInst
}

func (o *MliInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	ex, result := (b.AsDSigned() * a.AsDSigned()).Split()
	o.B.Write(state, result)
	state.WriteEX(ex)
	return nil
}

func (o *MliInst) Clone() Instruction {
	return &MliInst{o.binaryInst.clone()}
}

func (o *MliInst) String() string {
	return o.binaryInst.format("MLI")
}

// 0x06: DIV b, a - sets b to b/a, sets EX to ((b<<16)/a)&0xffff. if a==0, sets
// b and EX to 0 instead. (treats b, a as unsigned)
type DivInst struct {
	binaryInst
}

func (o *DivInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if a == 0 {
		o.B.Write(state, 0)
		state.WriteEX(0)
	} else {
		result, ex := ((DWord(b) << 16) / DWord(a)).Split()
		o.B.Write(state, result)
		state.WriteEX(ex)
	}
	return nil
}

func (o *DivInst) Clone() Instruction {
	return &DivInst{o.binaryInst.clone()}
}

func (o *DivInst) String() string {
	return o.binaryInst.format("DIV")
}

// 0x07: DVI b, a - like DIV, but treat b, a as signed. Rounds towards 0
type DviInst struct {
	binaryInst
}

func (o *DviInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if a == 0 {
		o.B.Write(state, 0)
		state.WriteEX(0)
	} else {
		result, ex := ((b.AsDSigned() << 16) / a.AsDSigned()).Split()
		o.B.Write(state, result)
		state.WriteEX(ex)
	}
	return nil
}

func (o *DviInst) Clone() Instruction {
	return &DviInst{o.binaryInst.clone()}
}

func (o *DviInst) String() string {
	return o.binaryInst.format("DVI")
}

// 0x08: MOD b, a - sets b to b%a. if a==0, sets b to 0 instructionead.
type ModInst struct {
	binaryInst
}

func (o *ModInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if a == 0 {
		o.B.Write(state, 0)
	} else {
		o.B.Write(state, b%a)
	}
	return nil
}

func (o *ModInst) Clone() Instruction {
	return &ModInst{o.binaryInst.clone()}
}

func (o *ModInst) String() string {
	return o.binaryInst.format("MOD")
}

// 0x09: MDI b, a - like MOD, but treat b, a as signed. (MDI -7, 16 == -7)
type MdiInst struct {
	binaryInst
}

func (o *MdiInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if a == 0 {
		o.B.Write(state, 0)
	} else {
		o.B.Write(state, Word(SWord(b)%SWord(a)))
	}
	return nil
}

func (o *MdiInst) Clone() Instruction {
	return &MdiInst{o.binaryInst.clone()}
}

func (o *MdiInst) String() string {
	return o.binaryInst.format("MDI")
}

// 0x0a: AND b, a - sets b to b&a
type AndInst struct {
	binaryInst
}

func (o *AndInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	o.B.Write(state, b&a)
	return nil
}

func (o *AndInst) Clone() Instruction {
	return &AndInst{o.binaryInst.clone()}
}

func (o *AndInst) String() string {
	return o.binaryInst.format("AND")
}

// 0x0b: BOR b, a - sets b to b|a
type BorInst struct {
	binaryInst
}

func (o *BorInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	o.B.Write(state, b|a)
	return nil
}

func (o *BorInst) Clone() Instruction {
	return &BorInst{o.binaryInst.clone()}
}

func (o *BorInst) String() string {
	return o.binaryInst.format("BOR")
}

// 0x0c: XOR b, a - sets b to b^a
type XorInst struct {
	binaryInst
}

func (o *XorInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	o.B.Write(state, b^a)
	return nil
}

func (o *XorInst) Clone() Instruction {
	return &XorInst{o.binaryInst.clone()}
}

func (o *XorInst) String() string {
	return o.binaryInst.format("XOR")
}

// 0x0d: SHR b, a - sets b to b>>a, sets EX to ((b<<16)>>a)&0xffff
type ShrInst struct {
	binaryInst
}

func (o *ShrInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result := (DWord(b) << 16) >> DWord(a)
	o.B.Write(state, Word(result>>16))
	state.WriteEX(Word(result & 0xffff))
	return nil
}

func (o *ShrInst) Clone() Instruction {
	return &ShrInst{o.binaryInst.clone()}
}

func (o *ShrInst) String() string {
	return o.binaryInst.format("SHR")
}

// 0x0e: ASR b, a - sets b to b>>a, sets EX to ((b<<16)>>>a)&0xffff (arithmetic
// shift) (treats b as signed)
type AsrInst struct {
	binaryInst
}

func (o *AsrInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result, ex := ((b.AsDSigned() << 16) >> a).Split()
	o.B.Write(state, result)
	state.WriteEX(ex)
	return nil
}

func (o *AsrInst) Clone() Instruction {
	return &AsrInst{o.binaryInst.clone()}
}

func (o *AsrInst) String() string {
	return o.binaryInst.format("ASR")
}

// 0x0f: SHL b, a - sets b to b<<a, sets EX to ((b<<a)>>16)&0xffff
type ShlInst struct {
	binaryInst
}

func (o *ShlInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	result := DWord(b) << DWord(a)
	o.B.Write(state, Word(result))
	state.WriteEX(Word(result >> 16))
	return nil
}

func (o *ShlInst) Clone() Instruction {
	return &ShlInst{o.binaryInst.clone()}
}

func (o *ShlInst) String() string {
	return o.binaryInst.format("SHL")
}

// 0x10: IFB b, a - performs next instruction only if (b&a)!=0
type IfbInst struct {
	binaryInst
}

func (o *IfbInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if (b & a) != 0 {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfbInst) Clone() Instruction {
	return &IfbInst{o.binaryInst.clone()}
}

func (o *IfbInst) String() string {
	return o.binaryInst.format("IFB")
}

// 0x11: IFC b, a - performs next instruction only if (b&a)==0
type IfcInst struct {
	binaryInst
}

func (o *IfcInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if (b & a) == 0 {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfcInst) Clone() Instruction {
	return &IfcInst{o.binaryInst.clone()}
}

func (o *IfcInst) String() string {
	return o.binaryInst.format("IFC")
}

// 0x12: IFE b, a - performs next instruction only if b==a
type IfeInst struct {
	binaryInst
}

func (o *IfeInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if b == a {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfeInst) Clone() Instruction {
	return &IfeInst{o.binaryInst.clone()}
}

func (o *IfeInst) String() string {
	return o.binaryInst.format("IFE")
}

// 0x13: IFN b, a - performs next instruction only if b!=a
type IfnInst struct {
	binaryInst
}

func (o *IfnInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if b != a {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfnInst) Clone() Instruction {
	return &IfnInst{o.binaryInst.clone()}
}

func (o *IfnInst) String() string {
	return o.binaryInst.format("IFN")
}

// 0x14: IFG b, a - performs next instruction only if b>a
type IfgInst struct {
	binaryInst
}

func (o *IfgInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if b > a {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfgInst) Clone() Instruction {
	return &IfgInst{o.binaryInst.clone()}
}

func (o *IfgInst) String() string {
	return o.binaryInst.format("IFG")
}

// 0x15: IFA b, a - performs next instruction only if b>a (signed)
type IfaInst struct {
	binaryInst
}

func (o *IfaInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if SWord(b) > SWord(a) {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfaInst) Clone() Instruction {
	return &IfaInst{o.binaryInst.clone()}
}

func (o *IfaInst) String() string {
	return o.binaryInst.format("IFA")
}

// 0x16: IFL b, a - performs next instruction only if b<a
type IflInst struct {
	binaryInst
}

func (o *IflInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if b < a {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IflInst) Clone() Instruction {
	return &IflInst{o.binaryInst.clone()}
}

func (o *IflInst) String() string {
	return o.binaryInst.format("IFL")
}

// 0x17: IFU b, a - performs next instruction only if b<a
type IfuInst struct {
	binaryInst
}

func (o *IfuInst) Execute(state MachineState) error {
	a, b := o.A.Read(state), o.B.Read(state)
	if b < a {
		return nil
	}
	return InstructionSkip(state, state)
}

func (o *IfuInst) Clone() Instruction {
	return &IfuInst{o.binaryInst.clone()}
}

func (o *IfuInst) String() string {
	return o.binaryInst.format("IFU")
}
