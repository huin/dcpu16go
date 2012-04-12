package model

import (
	"fmt"
)

type Operation interface {
	LoadNextWords(WordLoader)
	Execute(Context) error // TODO Return ticks.
	String() string
}

func OperationSkip(wordLoader WordLoader) error {
	// TODO Efficient version that doesn't hit memory?
	op, err := operationFromWord(wordLoader.WordLoad())
	if err != nil {
		return err
	}
	op.LoadNextWords(wordLoader)
	return nil
}

func OperationLoad(wordLoader WordLoader) (Operation, error) {
	op, err := operationFromWord(wordLoader.WordLoad())
	if err != nil {
		return nil, err
	}
	op.LoadNextWords(wordLoader)
	return op, err
}

func operationFromWord(w Word) (Operation, error) {
	opCode := w & 0x000f
	if opCode == 0x0 {
		// Non-basic instruction.
		extOpCode := (w >> 4) & 0x003f
		a := ValueFromWord((w >> 10) & 0x003f)
		switch extOpCode {
		case 0x00: // reserved for future expansion
		case 0x01: // JSR a - pushes the address of the next instruction to the stack, then sets PC to a
			return &JsrOp{unaryOp{a}}, nil
		default: // 0x02-0x3f: reserved
		}
		return nil, fmt.Errorf("unknown/reserved extOpCode 0x%02x", extOpCode)
	} else {
		a := ValueFromWord((w >> 4) & 0x003f)
		b := ValueFromWord((w >> 10) & 0x003f)
		switch opCode {
		case 0x1: // SET a, b - sets a to b
			return &SetOp{binaryOp{a, b}}, nil
		case 0x2: // ADD a, b - sets a to a+b, sets O to 0x0001 if there's an overflow, 0x0 otherwise
			return &AddOp{binaryOp{a, b}}, nil
		case 0x3: // SUB a, b - sets a to a-b, sets O to 0xffff if there's an underflow, 0x0 otherwise
			return &SubOp{binaryOp{a, b}}, nil
		case 0x4: // MUL a, b - sets a to a*b, sets O to ((a*b)>>16)&0xffff
			return &MulOp{binaryOp{a, b}}, nil
		case 0x5: // DIV a, b - sets a to a/b, sets O to ((a<<16)/b)&0xffff. if b==0, sets a and O to 0 instead.
			return &DivOp{binaryOp{a, b}}, nil
		case 0x6: // MOD a, b - sets a to a%b. if b==0, sets a to 0 instead.
			return &ModOp{binaryOp{a, b}}, nil
		case 0x7: // SHL a, b - sets a to a<<b, sets O to ((a<<b)>>16)&0xffff
			return &ShlOp{binaryOp{a, b}}, nil
		case 0x8: // SHR a, b - sets a to a>>b, sets O to ((a<<16)>>b)&0xffff
			return &ShrOp{binaryOp{a, b}}, nil
		case 0x9: // AND a, b - sets a to a&b
			return &AndOp{binaryOp{a, b}}, nil
		case 0xa: // BOR a, b - sets a to a|b
			return &BorOp{binaryOp{a, b}}, nil
		case 0xb: // XOR a, b - sets a to a^b
			return &XorOp{binaryOp{a, b}}, nil
		case 0xc: // IFE a, b - performs next instruction only if a==b
			return &IfeOp{binaryOp{a, b}}, nil
		case 0xd: // IFN a, b - performs next instruction only if a!=b
			return &IfnOp{binaryOp{a, b}}, nil
		case 0xe: // IFG a, b - performs next instruction only if a>b
			return &IfgOp{binaryOp{a, b}}, nil
		case 0xf: // IFB a, b - performs next instruction only if (a&b)!=0
			return &IfbOp{binaryOp{a, b}}, nil
		}
	}
	panic(fmt.Errorf("Operation code 0x%x out of range", opCode))
}

// unaryOp forms common data and code for operations that take one value.
type unaryOp struct {
	A Value
}

func (o *unaryOp) LoadNextWords(wordLoader WordLoader) {
	o.A.LoadOpValue(wordLoader)
}

func (o *unaryOp) format(name string) string {
	return fmt.Sprintf("%s %v", name, o.A)
}

// 0x01: JSR a - pushes the address of the next instruction to the stack, then sets PC to a
type JsrOp struct {
	unaryOp
}

func (o *JsrOp) Execute(ctx Context) error {
	ctx.WriteMemory(ctx.DecReadSP(), ctx.PC())
	ctx.WritePC(o.A.Read(ctx))
	return nil
}

func (o *JsrOp) String() string {
	return o.unaryOp.format("JSR")
}

// binaryOp forms common data and code for operations that take two values (A
// and B).
type binaryOp struct {
	A, B Value
}

func (o *binaryOp) LoadNextWords(wordLoader WordLoader) {
	o.A.LoadOpValue(wordLoader)
	o.B.LoadOpValue(wordLoader)
}

func (o *binaryOp) format(name string) string {
	return fmt.Sprintf("%s %v, %v", name, o.A, o.B)
}

// 0x1: SET a, b - sets a to b
type SetOp struct {
	binaryOp
}

func (o *SetOp) Execute(ctx Context) error {
	o.A.Write(ctx, o.B.Read(ctx))
	return nil
}

func (o *SetOp) String() string {
	return o.binaryOp.format("SET")
}

// 0x2: ADD a, b - sets a to a+b, sets O to 0x0001 if there's an overflow, 0x0 otherwise
type AddOp struct {
	binaryOp
}

func (o *AddOp) Execute(ctx Context) error {
	a, b := o.A.Read(ctx), o.B.Read(ctx)
	result := uint32(a) + uint32(b)
	o.A.Write(ctx, Word(result&0xffff))
	ctx.WriteO(Word(result >> 16))
	return nil
}

func (o *AddOp) String() string {
	return o.binaryOp.format("ADD")
}

// 0x3: SUB a, b - sets a to a-b, sets O to 0xffff if there's an underflow, 0x0 otherwise
type SubOp struct {
	binaryOp
}

func (o *SubOp) Execute(ctx Context) error {
	a, b := o.A.Read(ctx), o.B.Read(ctx)
	result := uint32(a) - uint32(b)
	o.A.Write(ctx, Word(result&0xffff))
	ctx.WriteO(Word(result >> 16))
	return nil
}

func (o *SubOp) String() string {
	return o.binaryOp.format("SUB")
}

// 0x4: MUL a, b - sets a to a*b, sets O to ((a*b)>>16)&0xffff
type MulOp struct {
	binaryOp
}

func (o *MulOp) Execute(ctx Context) error {
	a, b := o.A.Read(ctx), o.B.Read(ctx)
	result := uint32(a) * uint32(b)
	o.A.Write(ctx, Word(result&0xffff))
	ctx.WriteO(Word(result >> 16))
	return nil
}

func (o *MulOp) String() string {
	return o.binaryOp.format("MUL")
}

// 0x5: DIV a, b - sets a to a/b, sets O to ((a<<16)/b)&0xffff. if b==0, sets a and O to 0 instead.
type DivOp struct {
	binaryOp
}

func (o *DivOp) Execute(ctx Context) error {
	a, b := o.A.Read(ctx), o.B.Read(ctx)
	result := (uint32(a) << 16) / uint32(b)
	o.A.Write(ctx, Word(result>>16))
	ctx.WriteO(Word(result & 0xffff))
	return nil
}

func (o *DivOp) String() string {
	return o.binaryOp.format("DIV")
}

// 0x6: MOD a, b - sets a to a%b. if b==0, sets a to 0 instead.
type ModOp struct {
	binaryOp
}

func (o *ModOp) Execute(ctx Context) error {
	// TODO
	panic("unimplemented")
}

func (o *ModOp) String() string {
	return o.binaryOp.format("MOD")
}

// 0x7: SHL a, b - sets a to a<<b, sets O to ((a<<b)>>16)&0xffff
type ShlOp struct {
	binaryOp
}

func (o *ShlOp) Execute(ctx Context) error {
	a, b := o.A.Read(ctx), o.B.Read(ctx)
	result := uint32(a) << uint32(b)
	o.A.Write(ctx, Word(result))
	ctx.WriteO(Word(result >> 16))
	return nil
}

func (o *ShlOp) String() string {
	return o.binaryOp.format("SHL")
}

// 0x8: SHR a, b - sets a to a>>b, sets O to ((a<<16)>>b)&0xffff
type ShrOp struct {
	binaryOp
}

func (o *ShrOp) Execute(ctx Context) error {
	a, b := o.A.Read(ctx), o.B.Read(ctx)
	result := (uint32(a) << 16) >> uint32(b)
	o.A.Write(ctx, Word(result>>16))
	ctx.WriteO(Word(result & 0xffff))
	return nil
}

func (o *ShrOp) String() string {
	return o.binaryOp.format("SHR")
}

// 0x9: AND a, b - sets a to a&b
type AndOp struct {
	binaryOp
}

func (o *AndOp) Execute(ctx Context) error {
	// TODO
	panic("unimplemented")
}

func (o *AndOp) String() string {
	return o.binaryOp.format("AND")
}

// 0xa: BOR a, b - sets a to a|b
type BorOp struct {
	binaryOp
}

func (o *BorOp) Execute(ctx Context) error {
	// TODO
	panic("unimplemented")
}

func (o *BorOp) String() string {
	return o.binaryOp.format("BOR")
}

// 0xb: XOR a, b - sets a to a^b
type XorOp struct {
	binaryOp
}

func (o *XorOp) Execute(ctx Context) error {
	// TODO
	panic("unimplemented")
}

func (o *XorOp) String() string {
	return o.binaryOp.format("XOR")
}

// 0xc: IFE a, b - performs next instruction only if a==b
type IfeOp struct {
	binaryOp
}

func (o *IfeOp) Execute(ctx Context) error {
	a, b := o.A.Read(ctx), o.B.Read(ctx)
	if a != b {
		return OperationSkip(ctx)
	}
	return nil
}

func (o *IfeOp) String() string {
	return o.binaryOp.format("IFE")
}

// 0xd: IFN a, b - performs next instruction only if a!=b
type IfnOp struct {
	binaryOp
}

func (o *IfnOp) Execute(ctx Context) error {
	a, b := o.A.Read(ctx), o.B.Read(ctx)
	if a == b {
		return OperationSkip(ctx)
	}
	return nil
}

func (o *IfnOp) String() string {
	return o.binaryOp.format("IFN")
}

// 0xe: IFG a, b - performs next instruction only if a>b
type IfgOp struct {
	binaryOp
}

func (o *IfgOp) Execute(ctx Context) error {
	// TODO
	panic("unimplemented")
}

func (o *IfgOp) String() string {
	return o.binaryOp.format("IFG")
}

// 0xf: IFB a, b - performs next instruction only if (a&b)!=0
type IfbOp struct {
	binaryOp
}

func (o *IfbOp) Execute(ctx Context) error {
	// TODO
	panic("unimplemented")
}

func (o *IfbOp) String() string {
	return o.binaryOp.format("IFB")
}
