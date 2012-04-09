package model

import "fmt"

type Value interface {
	Write(Context, Word)
	Read(Context) Word
	LoadOpValue(WordLoader)
	String() string
}

// TODO: Optimize value reading by containing each type of value in a slice,
// and have ValueFromWord return (optionally reset) values from the slice. Have
// two slices per CPU, one for A and one for B values. This will cut out memory
// allocation/GCing.

func ValueFromWord(w Word) Value {
	switch {
	case 0x00 <= w && w <= 0x07:
		return &RegisterValue{Reg: RegisterId(w)}
	case 0x08 <= w && w <= 0x0f:
		return &RegisterAddressValue{Reg: RegisterId(w - 0x08)}
	case 0x10 <= w && w <= 0x17:
		return &RegisterRelAddressValue{Reg: RegisterId(w - 0x10)}
	case 0x18 == w:
		return PopValue{}
	case 0x19 == w:
		return PeekValue{}
	case 0x1a == w:
		return PushValue{}
	case 0x1b == w:
		return SpValue{}
	case 0x1c == w:
		return PcValue{}
	case 0x1d == w:
		return OValue{}
	case 0x1e == w:
		return &AddressValue{}
	case 0x1f == w:
		return &WordValue{}
	case 0x20 <= w && w <= 0x3f:
		return &LiteralValue{Literal: w - 0x20}
	}
	panic(fmt.Errorf("Value code 0x%02x out of range", w))
}

type noExtraWord struct{}

func (v noExtraWord) LoadOpValue(WordLoader) {
}

type extraWord struct {
	Value Word
}

func (v *extraWord) LoadOpValue(wordLoader WordLoader) {
	v.Value = wordLoader.WordLoad()
}

// 0x00-0x07: register
type RegisterValue struct {
	noExtraWord
	Reg RegisterId
}

func (v *RegisterValue) Write(ctx Context, word Word) {
	ctx.WriteRegister(v.Reg, word)
}

func (v *RegisterValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v *RegisterValue) String() string {
	return v.Reg.String()
}

// 0x08-0x0f: [register]
type RegisterAddressValue struct {
	noExtraWord
	Reg RegisterId
}

func (v *RegisterAddressValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v *RegisterAddressValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v *RegisterAddressValue) String() string {
	return "[" + v.Reg.String() + "]"
}

// 0x10-0x17: [next word + register]
type RegisterRelAddressValue struct {
	extraWord
	Reg RegisterId
}

func (v *RegisterRelAddressValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v *RegisterRelAddressValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v *RegisterRelAddressValue) String() string {
	return fmt.Sprintf("[0x%04x+%v]", v.Value, v.Reg)
}

// 0x18: POP / [SP++]
type PopValue struct {
	noExtraWord
}

func (v PopValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v PopValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v PopValue) String() string {
	return "POP"
}

// 0x19: PEEK / [SP]
type PeekValue struct {
	noExtraWord
}

func (v PeekValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v PeekValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v PeekValue) String() string {
	return "PEEK"
}

// 0x1a: PUSH / [--SP]
type PushValue struct {
	noExtraWord
}

func (v PushValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v PushValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v PushValue) String() string {
	return "PUSH"
}

// 0x1b: SP
type SpValue struct {
	noExtraWord
}

func (v SpValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v SpValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v SpValue) String() string {
	return "SP"
}

// 0x1c: PC
type PcValue struct {
	noExtraWord
}

func (v PcValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v PcValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v PcValue) String() string {
	return "PC"
}

// 0x1d: O
type OValue struct {
	noExtraWord
}

func (v OValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v OValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v OValue) String() string {
	return "O"
}

// 0x1e: [next word]
type AddressValue struct {
	extraWord
}

func (v *AddressValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v *AddressValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v *AddressValue) String() string {
	return fmt.Sprintf("[0x%04x]", v.Value)
}

// 0x1f: next word (literal)
type WordValue struct {
	extraWord
}

func (v *WordValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v *WordValue) Read(ctx Context) Word {
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

func (v *LiteralValue) Write(ctx Context, word Word) {
	// TODO
	panic("unimplemented")
}

func (v *LiteralValue) Read(ctx Context) Word {
	// TODO
	panic("unimplemented")
	return 0
}

func (v *LiteralValue) String() string {
	return fmt.Sprintf("0x%02x", v.Literal)
}
