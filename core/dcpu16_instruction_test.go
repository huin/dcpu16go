package core

import (
	"testing"
)

var basicInstructionSetImplTest InstructionSet = &D16InstructionSet{}

type FakeWordLoader struct {
	t     *testing.T
	Words []Word
	Loc   int
}

func (l *FakeWordLoader) WordLoad() (Word, error) {
	if l.Loc >= len(l.Words) {
		l.t.Fatalf("exceeded bounds on %#v", *l)
	}
	w := l.Words[l.Loc]
	l.Loc++
	return w, nil
}

func (l *FakeWordLoader) SkipWords(count Word) error {
	l.Loc += int(count)
	if l.Loc >= len(l.Words) {
		l.t.Fatalf("exceeded bounds on %#v", *l)
	}
	return nil
}

func (l *FakeWordLoader) exhausted() bool {
	return l.Loc == len(l.Words)
}

func TestInstructionLoad(t *testing.T) {
	var instructionSet D16InstructionSet

	for _, test := range notchExample {
		wordLoader := &FakeWordLoader{t, test.Words, 0}
		instruction, err := InstructionLoad(wordLoader, &instructionSet)
		if err != nil {
			t.Errorf("Instruction %#v returned error %v", test.Words, err)
			continue
		}
		if !wordLoader.exhausted() {
			t.Errorf("Instruction %v (%#v) did not exhaust words to load", instruction, test.Words)
			continue
		}
		str := instruction.String()
		if test.Str != str {
			t.Errorf("Instruction %v (%#v) disagrees on string repr (expected %q, got %q)",
				instruction, test.Words, test.Str, str)
		}
	}
}

// Returns a binaryInst with register A for value B, and the given literal for
// value A.
func binInstValue(literal Word) binaryInst {
	a := &WordValue{}
	a.Value = literal
	b := &RegisterValue{Reg: RegA}
	return binaryInst{a, b}
}

func TestInstructionExecute(t *testing.T) {
	type Test struct {
		Name   string
		InitA  Word
		InitEX Word
		Inst   Instruction
		ExpA   Word
		ExpEX  Word
	}

	tests := []Test{
		// AddInst
		{
			"ADD 0 + 5 = 5",
			0, 0, &AddInst{binInstValue(5)},
			5, 0,
		},
		{
			"ADD 0xffff + 0x0001 = 0x0000, with EX=0x0001",
			0xffff, 0, &AddInst{binInstValue(0x0001)},
			0x0000, 0x0001,
		},
		// SubInst
		{
			"SUB 10 - 2 = 8",
			10, 0, &SubInst{binInstValue(2)},
			8, 0,
		},
		{
			"SUB 0x000d - 0x000f = 0xfffe, with EX=0xffff",
			0x000d, 0, &SubInst{binInstValue(0x000f)},
			0xfffe, 0xffff,
		},
		// MulInst
		{
			"MUL 5 * 9 = 45",
			5, 0, &MulInst{binInstValue(9)},
			45, 0,
		},
		{
			"MUL 0x1234 * 0x0100 = 0x3400, with EX=0x0012",
			0x1234, 0, &MulInst{binInstValue(0x0100)},
			0x3400, 0x0012,
		},
		// MliInst
		{
			"MLI -5 * 9 = -45",
			Signed(-5), 0, &MliInst{binInstValue(9)},
			Signed(-45), 0xffff,
		},
		{
			"MLI 0x1234 * 0x0100 = 0x3400, with EX=0x0012",
			0x1234, 0, &MliInst{binInstValue(0x0100)},
			0x3400, 0x0012,
		},
		// DivInst
		{
			"DIV 12 / 3 = 4",
			12, 0, &DivInst{binInstValue(3)},
			4, 0,
		},
		{
			"DIV 12 / 3 = 4",
			12, 0, &DivInst{binInstValue(3)},
			4, 0,
		},
		{
			"DIV 0x1234 / 0x0100 = 0x0012, with EX=0x3400",
			0x1234, 0, &DivInst{binInstValue(0x0100)},
			0x0012, 0x3400,
		},
		{
			"DIV 5 / 0 = 0, with EX=0",
			5, 0, &DivInst{binInstValue(0)},
			0, 0,
		},
		// DviInst
		{
			"DVI 12 / -3 = -4",
			12, 0, &DviInst{binInstValue(Signed(-3))},
			Signed(-4), 0,
		},
		{
			"DVI -12 / 3 = -4",
			Signed(-12), 0, &DviInst{binInstValue(3)},
			Signed(-4), 0,
		},
		{
			"DVI -12 / -3 = 4",
			Signed(-12), 0, &DviInst{binInstValue(Signed(-3))},
			4, 0,
		},
		{
			"DVI 0x1234 / 0x0100 = 0x0012, with EX=0x3400",
			0x1234, 0, &DviInst{binInstValue(0x0100)},
			0x0012, 0x3400,
		},
		{
			"DVI 5 / 0 = 0, with EX=0",
			5, 0, &DviInst{binInstValue(0)},
			0, 0,
		},
		// ModInst
		{
			"MOD 12 % 3 = 0",
			12, 0, &ModInst{binInstValue(3)},
			0, 0,
		},
		{
			"MOD 12 % 5 = 2",
			12, 0, &ModInst{binInstValue(5)},
			2, 0,
		},
		{
			"MOD 0x1234 % 0x0100 = 0x0034",
			0x1234, 0, &ModInst{binInstValue(0x0100)},
			0x0034, 0,
		},
		{
			"MOD 5 % 0 = 0",
			5, 0, &ModInst{binInstValue(0)},
			0, 0,
		},
		// MdiInst
		{
			"MDI 12 % 3 = 0",
			12, 0, &MdiInst{binInstValue(3)},
			0, 0,
		},
		{
			"MDI 12 % 5 = 2",
			12, 0, &MdiInst{binInstValue(5)},
			2, 0,
		},
		{
			"MDI -7, 16 = -7",
			Signed(-7), 0, &MdiInst{binInstValue(16)},
			Signed(-7), 0,
		},
		{
			"MDI 0x1234 % 0x0100 = 0x0034",
			0x1234, 0, &MdiInst{binInstValue(0x0100)},
			0x0034, 0,
		},
		{
			"MDI 5 % 0 = 0",
			5, 0, &MdiInst{binInstValue(0)},
			0, 0,
		},
		// AndInst
		{
			"AND 0x1230 & 0x0410 = 0x0010",
			0x1230, 0, &AndInst{binInstValue(0x0410)},
			0x0010, 0x0000,
		},
		// BorInst
		{
			"BOR 0x1230 | 0x0410 = 0x1630",
			0x1230, 0, &BorInst{binInstValue(0x0410)},
			0x1630, 0x0000,
		},
		// XorInst
		{
			"XOR 0x1230 ^ 0x0410 = 0x1620",
			0x1230, 0, &XorInst{binInstValue(0x0410)},
			0x1620, 0x0000,
		},
		// ShrInst
		{
			"SHR 0x1230 >> 4 = 0x0123",
			0x1230, 0, &ShrInst{binInstValue(4)},
			0x0123, 0x0000,
		},
		{
			"SHR 0x1230 >> 8 = 0x0001, with EX=0x2300",
			0x0123, 0, &ShrInst{binInstValue(8)},
			0x0001, 0x2300,
		},
		// AsrInst
		{
			"ASR 0x1230 >>> 4 = 0x0123",
			0x1230, 0, &AsrInst{binInstValue(4)},
			0x0123, 0x0000,
		},
		{
			"ASR -5 >>> 1 = -3, with EX=0x8000",
			Signed(-5), 0, &AsrInst{binInstValue(1)},
			Signed(-3), 0x8000,
		},
		// ShlInst
		{
			"SHL 0x0123 << 4 = 0x1230",
			0x0123, 0, &ShlInst{binInstValue(4)},
			0x1230, 0x0000,
		},
		{
			"SHL 0x0123 << 12 = 0x3000, with EX=0x0012",
			0x0123, 0, &ShlInst{binInstValue(12)},
			0x3000, 0x0012,
		},
		// AdxInst
		{
			"ADX 2, 1 (EX=3) = 6, with EX = 0",
			1, 3, &AdxInst{binInstValue(2)},
			6, 0,
		},
		{
			"ADX 0x7fff, 0 (EX=0x8000) = 0xffff, with EX = 0",
			0x7fff, 0x8000, &AdxInst{binInstValue(0)},
			0xffff, 0,
		},
		{
			"ADX 0x7fff, 1 (EX=0x8000) = 0, with EX = 0x0001",
			0x7fff, 0x8000, &AdxInst{binInstValue(1)},
			0, 0x0001,
		},
		// SbxInst
		{
			"SBX 0xffff, 0 (EX=0xffff) = 0xfffe, with EX = 0",
			0xffff, 0xffff, &SbxInst{binInstValue(0)},
			0xfffe, 0,
		},
	}

	for _, test := range tests {
		var state D16MachineState
		state.Init()
		state.D16CPU.WriteRegister(RegA, test.InitA)
		state.D16CPU.WriteEX(test.InitEX)
		test.Inst.Execute(&state)
		if test.ExpA != state.D16CPU.registers[RegA] || test.ExpEX != state.D16CPU.ex {
			t.Errorf("%s: %v", test.Name, test.Inst.String())
			t.Errorf("  expected: A=0x%04x EX=0x%04x", test.ExpA, test.ExpEX)
			t.Errorf("       got: A=0x%04x EX=0x%04x", state.D16CPU.registers[RegA], state.D16CPU.ex)
		}
	}
}
