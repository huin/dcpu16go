package model

import (
	"testing"
)

// Returns a binaryInst with value A as register A, and value B as the given
// literal value.
func binInstValue(bValue Word) binaryInst {
	a := &RegisterValue{Reg: RegA}
	b := &WordValue{}
	b.Value = bValue
	return binaryInst{a, b}
}

func TestInstructionExecute(t *testing.T) {
	type Test struct {
		Name  string
		InitA Word
		Inst  Instruction
		ExpA  Word
		ExpO  Word
	}

	tests := []Test{
		// AddInst
		{
			"ADD 0 + 5 = 5",
			0, &AddInst{binInstValue(5)},
			5, 0,
		},
		{
			"ADD 0xffff + 0x0001 = 0x0000, carry=0x0001",
			0xffff, &AddInst{binInstValue(0x0001)},
			0x0000, 0x0001,
		},
		// SubInst
		{
			"SUB 10 - 2 = 8",
			10, &SubInst{binInstValue(2)},
			8, 0,
		},
		{
			"SUB 0x000d - 0x000f = 0xfffe, with carry=0xffff",
			0x000d, &SubInst{binInstValue(0x000f)},
			0xfffe, 0xffff,
		},
		// MulInst
		{
			"MUL 5 * 9 = 45",
			5, &MulInst{binInstValue(9)},
			45, 0,
		},
		{
			"MUL 0x1234 * 0x0100 = 0x3400, with carry=0x0012",
			0x1234, &MulInst{binInstValue(0x0100)},
			0x3400, 0x0012,
		},
		// DivInst
		{
			"DIV 12 / 3 = 4",
			12, &DivInst{binInstValue(3)},
			4, 0,
		},
		{
			"DIV 0x1234 / 0x0100 = 0x0012, with carry=0x3400",
			0x1234, &DivInst{binInstValue(0x0100)},
			0x0012, 0x3400,
		},
		// ModInst
		{
			"MOD 12 % 3 = 0",
			12, &ModInst{binInstValue(3)},
			0, 0,
		},
		{
			"MOD 0x1234 % 0x0100 = 0x0034, with carry=0x0000",
			0x1234, &ModInst{binInstValue(0x0100)},
			0x0034, 0x0000,
		},
		// ShlInst
		{
			"SHL 0x0123 << 4 = 0x1230",
			0x0123, &ShlInst{binInstValue(4)},
			0x1230, 0x0000,
		},
		{
			"SHL 0x0123 << 12 = 0x3000, with carry=0x0012",
			0x0123, &ShlInst{binInstValue(12)},
			0x3000, 0x0012,
		},
		// ShrInst
		{
			"SHR 0x1230 >> 4 = 0x0123",
			0x1230, &ShrInst{binInstValue(4)},
			0x0123, 0x0000,
		},
		{
			"SHR 0x1230 >> 8 = 0x0001, with carry=0x2300",
			0x0123, &ShrInst{binInstValue(8)},
			0x0001, 0x2300,
		},
		// AndInst
		{
			"AND 0x1230 & 0x0410 = 0x0010",
			0x1230, &AndInst{binInstValue(0x0410)},
			0x0010, 0x0000,
		},
		// BorInst
		{
			"BOR 0x1230 | 0x0410 = 0x1630",
			0x1230, &BorInst{binInstValue(0x0410)},
			0x1630, 0x0000,
		},
		// XorInst
		{
			"XOR 0x1230 ^ 0x0410 = 0x1620",
			0x1230, &XorInst{binInstValue(0x0410)},
			0x1620, 0x0000,
		},
	}

	for _, test := range tests {
		var state BasicMachineState
		state.Init()
		state.BasicCPU.WriteRegister(RegA, test.InitA)
		test.Inst.Execute(&state)
		if test.ExpA != state.BasicCPU.registers[RegA] || test.ExpO != state.BasicCPU.o {
			t.Errorf("%s: %v", test.Name, test.Inst.String())
			t.Errorf("  expected: A=0x%04x O=0x%04x", test.ExpA, test.ExpO)
			t.Errorf("       got: A=0x%04x O=0x%04x", state.BasicCPU.registers[RegA], state.BasicCPU.o)
		}
	}
}
