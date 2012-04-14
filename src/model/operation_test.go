package model

import (
	"testing"
)

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

func (l *FakeWordLoader) exhausted() bool {
	return l.Loc == len(l.Words)
}

func TestOperationLoad(t *testing.T) {
	type Test struct {
		Expected string
		Words    []Word
	}

	// Test examples taken from dcpu-16.txt, and modified such that literal lengths
	// indicate "next word" (4 hex digits) vs "embedded literal" (2 hex digits),
	// and replacing labels with address values.
	tests := []Test{
		// Try some basic stuff
		{"SET A, 0x0030", []Word{0x7c01, 0x0030}},
		{"SET [0x1000], 0x0020", []Word{0x7de1, 0x1000, 0x0020}},
		{"SUB A, [0x1000]", []Word{0x7803, 0x1000}},
		{"IFN A, 0x10", []Word{0xc00d}},
		{"SET PC, 0x001a", []Word{0x7dc1, 0x001a}},
		// Do a loopy thing
		{"SET I, 0x0a", []Word{0xa861}},
		{"SET A, 0x2000", []Word{0x7c01, 0x2000}},
		{"SET [0x2000+I], [A]", []Word{0x2161, 0x2000}},
		{"SUB I, 0x01", []Word{0x8463}},
		{"IFN I, 0x00", []Word{0x806d}},
		{"SET PC, 0x000d", []Word{0x7dc1, 0x000d}},
		// Call a subroutine
		{"SET X, 0x04", []Word{0x9031}},
		{"JSR 0x0018", []Word{0x7c10, 0x0018}},
		{"SET PC, 0x001a", []Word{0x7dc1, 0x001a}},
		{"SHL X, 0x04", []Word{0x9037}},
		{"SET PC, POP", []Word{0x61c1}},
		// Hang forever. X should now be 0x40 if everything went right.
		{"SET PC, 0x001a", []Word{0x7dc1, 0x001a}},
	}

	for _, test := range tests {
		wordLoader := &FakeWordLoader{t, test.Words, 0}
		op, err := OperationLoad(wordLoader)
		if err != nil {
			t.Errorf("Operation %#v returned error %v", test.Words, err)
			continue
		}
		if !wordLoader.exhausted() {
			t.Errorf("Operation %v (%#v) did not exaust words to load", op, test.Words)
			continue
		}
		str := op.String()
		if test.Expected != str {
			t.Errorf("Operation %v (%#v) disagrees on string repr (expected %q, got %q)",
				op, test.Words, test.Expected, str)
		}
	}
}

// Returns a binaryOp with value A as register A, and value B as the given
// literal value.
func binOpValue(bValue Word) binaryOp {
	a := &RegisterValue{Reg: RegA}
	b := &WordValue{}
	b.Value = bValue
	return binaryOp{a, b}
}

func TestOperationExecute(t *testing.T) {
	type Test struct {
		Name  string
		InitA Word
		Op    Operation
		ExpA  Word
		ExpO  Word
	}

	tests := []Test{
		// AddOp
		{
			"ADD 0 + 5 = 5",
			0, &AddOp{binOpValue(5)},
			5, 0,
		},
		{
			"ADD 0xffff + 0x0001 = 0x0000, carry=0x0001",
			0xffff, &AddOp{binOpValue(0x0001)},
			0x0000, 0x0001,
		},
		// SubOp
		{
			"SUB 10 - 2 = 8",
			10, &SubOp{binOpValue(2)},
			8, 0,
		},
		{
			"SUB 0x000d - 0x000f = 0xfffe, with carry=0xffff",
			0x000d, &SubOp{binOpValue(0x000f)},
			0xfffe, 0xffff,
		},
		// MulOp
		{
			"MUL 5 * 9 = 45",
			5, &MulOp{binOpValue(9)},
			45, 0,
		},
		{
			"MUL 0x1234 * 0x0100 = 0x3400, with carry=0x0012",
			0x1234, &MulOp{binOpValue(0x0100)},
			0x3400, 0x0012,
		},
		// DivOp
		{
			"DIV 12 / 3 = 4",
			12, &DivOp{binOpValue(3)},
			4, 0,
		},
		{
			"DIV 0x1234 / 0x0100 = 0x0012, with carry=0x3400",
			0x1234, &DivOp{binOpValue(0x0100)},
			0x0012, 0x3400,
		},
		// ModOp
		{
			"MOD 12 % 3 = 0",
			12, &ModOp{binOpValue(3)},
			0, 0,
		},
		{
			"MOD 0x1234 % 0x0100 = 0x0034, with carry=0x0000",
			0x1234, &ModOp{binOpValue(0x0100)},
			0x0034, 0x0000,
		},
		// ShlOp
		{
			"SHL 0x0123 << 4 = 0x1230",
			0x0123, &ShlOp{binOpValue(4)},
			0x1230, 0x0000,
		},
		{
			"SHL 0x0123 << 12 = 0x3000, with carry=0x0012",
			0x0123, &ShlOp{binOpValue(12)},
			0x3000, 0x0012,
		},
		// ShrOp
		{
			"SHR 0x1230 >> 4 = 0x0123",
			0x1230, &ShrOp{binOpValue(4)},
			0x0123, 0x0000,
		},
		{
			"SHR 0x1230 >> 8 = 0x0001, with carry=0x2300",
			0x0123, &ShrOp{binOpValue(8)},
			0x0001, 0x2300,
		},
		// AndOp
		{
			"AND 0x1230 & 0x0410 = 0x0010",
			0x1230, &AndOp{binOpValue(0x0410)},
			0x0010, 0x0000,
		},
		// BorOp
		{
			"BOR 0x1230 | 0x0410 = 0x1630",
			0x1230, &BorOp{binOpValue(0x0410)},
			0x1630, 0x0000,
		},
		// XorOp
		{
			"XOR 0x1230 ^ 0x0410 = 0x1620",
			0x1230, &XorOp{binOpValue(0x0410)},
			0x1620, 0x0000,
		},
	}

	for _, test := range tests {
		var state BasicMachineState
		state.Init()
		state.BasicCPU.WriteRegister(RegA, test.InitA)
		test.Op.Execute(&state)
		if test.ExpA != state.BasicCPU.registers[RegA] || test.ExpO != state.BasicCPU.o {
			t.Errorf("%s: %v", test.Name, test.Op.String())
			t.Errorf("  expected: A=0x%04x O=0x%04x", test.ExpA, test.ExpO)
			t.Errorf("       got: A=0x%04x O=0x%04x", state.BasicCPU.registers[RegA], state.BasicCPU.o)
		}
	}
}
