package model

import (
	"fmt"
)

type BasicInstructionSet struct {
}

func (is *BasicInstructionSet) Init() {
}

func (is *BasicInstructionSet) Instruction(w Word) (Instruction, error) {
	opCode := w & 0x000f
	if opCode == 0x0 {
		// Non-basic instruction.
		extOpCode := (w >> 4) & 0x003f
		a := ValueFromWord((w >> 10) & 0x003f)
		switch extOpCode {
		case 0x00: // reserved for future expansion
		case 0x01: // JSR a - pushes the address of the next instruction to the stack, then sets PC to a
			return &JsrInst{unaryInst{a}}, nil
		default: // 0x02-0x3f: reserved
		}
		return nil, fmt.Errorf("unknown/reserved extOpCode 0x%02x", extOpCode)
	} else {
		a := ValueFromWord((w >> 4) & 0x003f)
		b := ValueFromWord((w >> 10) & 0x003f)
		switch opCode {
		case 0x1: // SET a, b - sets a to b
			return &SetInst{binaryInst{a, b}}, nil
		case 0x2: // ADD a, b - sets a to a+b, sets O to 0x0001 if there's an overflow, 0x0 otherwise
			return &AddInst{binaryInst{a, b}}, nil
		case 0x3: // SUB a, b - sets a to a-b, sets O to 0xffff if there's an underflow, 0x0 otherwise
			return &SubInst{binaryInst{a, b}}, nil
		case 0x4: // MUL a, b - sets a to a*b, sets O to ((a*b)>>16)&0xffff
			return &MulInst{binaryInst{a, b}}, nil
		case 0x5: // DIV a, b - sets a to a/b, sets O to ((a<<16)/b)&0xffff. if b==0, sets a and O to 0 instructionead.
			return &DivInst{binaryInst{a, b}}, nil
		case 0x6: // MOD a, b - sets a to a%b. if b==0, sets a to 0 instructionead.
			return &ModInst{binaryInst{a, b}}, nil
		case 0x7: // SHL a, b - sets a to a<<b, sets O to ((a<<b)>>16)&0xffff
			return &ShlInst{binaryInst{a, b}}, nil
		case 0x8: // SHR a, b - sets a to a>>b, sets O to ((a<<16)>>b)&0xffff
			return &ShrInst{binaryInst{a, b}}, nil
		case 0x9: // AND a, b - sets a to a&b
			return &AndInst{binaryInst{a, b}}, nil
		case 0xa: // BOR a, b - sets a to a|b
			return &BorInst{binaryInst{a, b}}, nil
		case 0xb: // XOR a, b - sets a to a^b
			return &XorInst{binaryInst{a, b}}, nil
		case 0xc: // IFE a, b - performs next instruction only if a==b
			return &IfeInst{binaryInst{a, b}}, nil
		case 0xd: // IFN a, b - performs next instruction only if a!=b
			return &IfnInst{binaryInst{a, b}}, nil
		case 0xe: // IFG a, b - performs next instruction only if a>b
			return &IfgInst{binaryInst{a, b}}, nil
		case 0xf: // IFB a, b - performs next instruction only if (a&b)!=0
			return &IfbInst{binaryInst{a, b}}, nil
		}
	}
	panic(fmt.Errorf("Instruction code 0x%x out of range", opCode))
}
