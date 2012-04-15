package core

import (
	"fmt"
)

// D16ValueSet is the table of basic values supported by the DCPU-16. This
// implementation avoids memory allocation, but the instruction returned is
// only good for use until the next call to the Value method.
type D16ValueSet struct {
	registerValue           RegisterValue
	registerAddressValue    RegisterAddressValue
	registerRelAddressValue RegisterRelAddressValue
	addressValue            AddressValue
	wordValue               WordValue
	literalValue            LiteralValue
}

func (vs *D16ValueSet) Value(w Word) (Value, error) {
	var value Value
	switch {
	case 0x00 <= w && w <= 0x07:
		vs.registerValue.Reg = RegisterId(w)
		return &vs.registerValue, nil
	case 0x08 <= w && w <= 0x0f:
		vs.registerAddressValue.Reg = RegisterId(w - 0x08)
		return &vs.registerAddressValue, nil
	case 0x10 <= w && w <= 0x17:
		vs.registerRelAddressValue.Reg = RegisterId(w - 0x10)
		return &vs.registerRelAddressValue, nil
	case 0x18 == w:
		value = PopValue{}
	case 0x19 == w:
		value = PeekValue{}
	case 0x1a == w:
		value = PushValue{}
	case 0x1b == w:
		value = SpValue{}
	case 0x1c == w:
		value = PcValue{}
	case 0x1d == w:
		value = OValue{}
	case 0x1e == w:
		value = &vs.addressValue
	case 0x1f == w:
		value = &vs.wordValue
	case 0x20 <= w && w <= 0x3f:
		vs.literalValue.Literal = w - 0x20
		return &vs.literalValue, nil
	default:
		return nil, ValueCodeError(w)
	}
	return value, nil
}

type noExtraWord struct{}

func (v noExtraWord) LoadInstValue(WordLoader) error {
	return nil
}

func (v noExtraWord) NumExtraWords() Word {
	return 0
}

type extraWord struct {
	Value Word
}

func (v *extraWord) LoadInstValue(wordLoader WordLoader) error {
	var err error
	v.Value, err = wordLoader.WordLoad()
	return err
}

func (v *extraWord) NumExtraWords() Word {
	return 1
}

// 0x00-0x07: register
type RegisterValue struct {
	noExtraWord
	Reg RegisterId
}

func (v *RegisterValue) Write(state MachineState, word Word) {
	state.WriteRegister(v.Reg, word)
}

func (v *RegisterValue) Read(state MachineState) Word {
	return state.Register(v.Reg)
}

func (v *RegisterValue) String() string {
	return v.Reg.String()
}

// 0x08-0x0f: [register]
type RegisterAddressValue struct {
	noExtraWord
	Reg RegisterId
}

func (v *RegisterAddressValue) Write(state MachineState, word Word) {
	state.WriteMemory(state.Register(v.Reg), word)
}

func (v *RegisterAddressValue) Read(state MachineState) Word {
	return state.ReadMemory(state.Register(v.Reg))
}

func (v *RegisterAddressValue) String() string {
	return "[" + v.Reg.String() + "]"
}

// 0x10-0x17: [next word + register]
type RegisterRelAddressValue struct {
	extraWord
	Reg RegisterId
}

func (v *RegisterRelAddressValue) Write(state MachineState, word Word) {
	state.WriteMemory(state.Register(v.Reg)+v.Value, word)
}

func (v *RegisterRelAddressValue) Read(state MachineState) Word {
	return state.ReadMemory(state.Register(v.Reg) + v.Value)
}

func (v *RegisterRelAddressValue) String() string {
	return fmt.Sprintf("[0x%04x+%v]", v.Value, v.Reg)
}

// 0x18: POP / [SP++]
type PopValue struct {
	noExtraWord
}

func (v PopValue) Write(state MachineState, word Word) {
	state.WriteMemory(state.ReadIncSP(), word)
}

func (v PopValue) Read(state MachineState) Word {
	return state.ReadMemory(state.ReadIncSP())
}

func (v PopValue) String() string {
	return "POP"
}

// 0x19: PEEK / [SP]
type PeekValue struct {
	noExtraWord
}

func (v PeekValue) Write(state MachineState, word Word) {
	state.WriteMemory(state.SP(), word)
}

func (v PeekValue) Read(state MachineState) Word {
	return state.ReadMemory(state.SP())
}

func (v PeekValue) String() string {
	return "PEEK"
}

// 0x1a: PUSH / [--SP]
type PushValue struct {
	noExtraWord
}

func (v PushValue) Write(state MachineState, word Word) {
	state.WriteMemory(state.DecReadSP(), word)
}

func (v PushValue) Read(state MachineState) Word {
	return state.ReadMemory(state.DecReadSP())
}

func (v PushValue) String() string {
	return "PUSH"
}

// 0x1b: SP
type SpValue struct {
	noExtraWord
}

func (v SpValue) Write(state MachineState, word Word) {
	state.WriteSP(word)
}

func (v SpValue) Read(state MachineState) Word {
	return state.SP()
}

func (v SpValue) String() string {
	return "SP"
}

// 0x1c: PC
type PcValue struct {
	noExtraWord
}

func (v PcValue) Write(state MachineState, word Word) {
	state.WritePC(word)
}

func (v PcValue) Read(state MachineState) Word {
	return state.PC()
}

func (v PcValue) String() string {
	return "PC"
}

// 0x1d: O
type OValue struct {
	noExtraWord
}

func (v OValue) Write(state MachineState, word Word) {
	state.WriteO(word)
}

func (v OValue) Read(state MachineState) Word {
	return state.O()
}

func (v OValue) String() string {
	return "O"
}

// 0x1e: [next word]
type AddressValue struct {
	extraWord
}

func (v *AddressValue) Write(state MachineState, word Word) {
	state.WriteMemory(v.extraWord.Value, word)
}

func (v *AddressValue) Read(state MachineState) Word {
	return state.ReadMemory(v.extraWord.Value)
}

func (v *AddressValue) String() string {
	return fmt.Sprintf("[0x%04x]", v.Value)
}

// 0x1f: next word (literal)
type WordValue struct {
	extraWord
}

func (v *WordValue) Write(state MachineState, word Word) {
	// No-op.
}

func (v *WordValue) Read(state MachineState) Word {
	return v.extraWord.Value
}

func (v *WordValue) String() string {
	return fmt.Sprintf("0x%04x", v.Value)
}

// 0x20-0x3f: literal value 0x00-0x1f (literal)
type LiteralValue struct {
	noExtraWord
	Literal Word
}

func (v *LiteralValue) Write(state MachineState, word Word) {
	// No-op.
}

func (v *LiteralValue) Read(state MachineState) Word {
	return v.Literal
}

func (v *LiteralValue) String() string {
	return fmt.Sprintf("0x%02x", v.Literal)
}
