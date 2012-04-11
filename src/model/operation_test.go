package model

import (
	"testing"
)

type FakeWordLoader struct {
	t     *testing.T
	Words []Word
	Loc   int
}

func (l *FakeWordLoader) WordLoad() Word {
	if l.Loc >= len(l.Words) {
		l.t.Fatalf("exceeded bounds on %#v", *l)
	}
	w := l.Words[l.Loc]
	l.Loc++
	return w
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
		Op    Operation
		InitA Word
		ExpA  Word
		ExpO  Word
	}

	tests := []Test{
		// AddOp
		{
			"ADD 0 + 5 = 5",
			&AddOp{binOpValue(5)},
			0,
			5, 0,
		},
		{
			"ADD 0xffff + 0x0001 = 0x0000, carry=0x0001",
			&AddOp{binOpValue(0x0001)},
			0xffff,
			0x0000, 0x0001,
		},
		// SubOp
		{
			"SUB 10 - 2 = 8",
			&SubOp{binOpValue(2)},
			10,
			8, 0,
		},
		{
			"SUB 0x000d - 0x000f = 0xfffe, with carry=0xffff",
			&SubOp{binOpValue(0x000f)},
			0x000d,
			0xfffe, 0xffff,
		},
	}

	for _, test := range tests {
		var ctx StandardContext
		ctx.Init()
		ctx.CPUState.WriteRegister(RegA, test.InitA)
		test.Op.Execute(&ctx)
		if test.ExpA != ctx.CPUState.registers[RegA] || test.ExpO != ctx.CPUState.o {
			t.Errorf("%s: %v", test.Name, test.Op.String())
			t.Errorf("  expected: A=0x%04x O=0x%04x", test.ExpA, test.ExpO)
			t.Errorf("       got: A=0x%04x O=0x%04x", ctx.CPUState.registers[RegA], ctx.CPUState.o)
		}
	}
}
