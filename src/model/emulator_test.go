package model

import (
	"testing"
)

type ExpMem struct {
	Offset   int
	Expected []Word
}

func (exp *ExpMem) CheckMem(t *testing.T, testName string, mem *MemoryState) bool {
	for i := range exp.Expected {
		if exp.Expected[i] != mem.Data[exp.Offset+i] {
			t.Errorf("%s: memory state at 0x%04x", testName, exp.Offset+i)
			t.Errorf("expected: %#v", exp.Expected[i])
			t.Errorf("got:      %#v", mem.Data[exp.Offset+i])
			return false
		}
	}
	return true
}

type ExpState struct {
	Name     string
	NumSteps int
	CPU      CPUState
	Mems     []ExpMem
}

func (exp *ExpState) CheckState(t *testing.T, testName string, ctx *StandardContext) bool {
	testName = testName + "::" + exp.Name

	for i := 0; i < exp.NumSteps; i++ {
		Step(ctx)
	}

	if !CPUEquals(&exp.CPU, &ctx.CPUState) {
		t.Errorf("%s: CPU state", testName)
		t.Errorf("expected: %#v", exp.CPU)
		t.Errorf("got:      %#v", ctx.CPUState)
		return false
	}

	for i := range exp.Mems {
		if !exp.Mems[i].CheckMem(t, testName, &ctx.MemoryState) {
			return false
		}
	}

	return true
}

var notchExample = []Word{
	// Try some basic stuff
	0x7c01, 0x0030, // SET A, 0x30
	0x7de1, 0x1000, 0x0020, // SET [0x1000], 0x20
	0x7803, 0x1000, // SUB A, [0x1000]
	0xc00d,         // IFN A, 0x10
	0x7dc1, 0x001a, // SET PC, crash

	// Do a loopy thing
	0xa861,         // SET I, 10
	0x7c01, 0x2000, // SET A, 0x2000
	0x2161, 0x2000, // :loop         SET [0x2000+I], [A]
	0x8463,         // SUB I, 1
	0x806d,         // IFN I, 0
	0x7dc1, 0x000d, // SET PC, loop

	// Call a subroutine
	0x9031,         // SET X, 0x4
	0x7c10, 0x0018, // JSR testsub
	0x7dc1, 0x001a, // SET PC, crash

	0x9037, // :testsub      SHL X, 4
	0x61c1, // SET PC, POP

	// Hang forever. X should now be 0x40 if everything went right.
	0x7dc1, 0x001a, // :crash        SET PC, crash
}

func TestStep(t *testing.T) {
	type Test struct {
		Name      string
		InitMem   []Word
		ExpStates []ExpState
	}

	tests := []Test{
		{
			Name:    "SET A, 0x0030",
			InitMem: []Word{0x7c01, 0x0030},
			ExpStates: []ExpState{
				ExpState{
					NumSteps: 1,
					CPU:      CPUState{registers: [8]Word{0x0030, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0002, sp: 0xffff},
					Mems:     []ExpMem{{0x0001, []Word{0x0030}}},
				},
			},
		},
		{
			Name:    "SET [0x0003], 0xbeef",
			InitMem: []Word{0x7de1, 0x0003, 0xbeef, 0x0000},
			ExpStates: []ExpState{
				ExpState{
					NumSteps: 1,
					CPU:      CPUState{registers: [8]Word{0, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0003, sp: 0xffff},
					Mems:     []ExpMem{{0x0003, []Word{0xbeef}}},
				},
			},
		},
		{
			Name: "SET A, 0x0030 / SET PUSH, A / SET C, PEEK / SET B, POP",
			InitMem: []Word{
				0x7c01, 0x0030, // SET A, 0x0030
				0x01a1, // SET PUSH, A
				0x6421, // SET C, PEEK
				0x6011, // SET B, POP
			},
			ExpStates: []ExpState{
				ExpState{
					NumSteps: 4,
					CPU:      CPUState{registers: [8]Word{0x0030, 0x0030, 0x0030, 0, 0, 0, 0, 0}, pc: 0x0005, sp: 0xffff},
					Mems:     []ExpMem{{0xfffe, []Word{0x0030, 0x0000}}},
				},
			},
		},
		{
			Name:    "Notch example",
			InitMem: notchExample,
			ExpStates: []ExpState{
				ExpState{
					Name:     "after basic stuff",
					NumSteps: 4,
					CPU:      CPUState{registers: [8]Word{0x0010}, pc: 0x000a, sp: 0xffff, o: 0x0000},
					Mems: []ExpMem{
						{0x1000, []Word{0x0020}},
					},
				},
				ExpState{
					Name:     "after loop init",
					NumSteps: 2,
					CPU:      CPUState{registers: [8]Word{0x2000, 0, 0, 0, 0, 0, 10, 0}, pc: 0x000d, sp: 0xffff, o: 0x0000},
				},
				ExpState{
					Name:     "after first test+iteration",
					NumSteps: 4,
					CPU:      CPUState{registers: [8]Word{0x2000, 0, 0, 0, 0, 0, 9, 0}, pc: 0x000d, sp: 0xffff, o: 0x0000},
					Mems: []ExpMem{
						{0x2010, []Word{0x0000}},
					},
				},
				ExpState{
					Name:     "after loop completion",
					NumSteps: 4*8 + 3,
					CPU:      CPUState{registers: [8]Word{0x2000, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0013, sp: 0xffff, o: 0x0000},
					Mems: []ExpMem{
						{0x2000, []Word{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000}},
					},
				},
				// TODO Additional state tests.
			},
		},
	}

	for i := range tests {
		test := &tests[i]

		ctx := StandardContext{}
		ctx.CPUState.Init()
		copy(ctx.MemoryState.Data[0:], test.InitMem)

		for expStateIdx := range test.ExpStates {
			expState := &test.ExpStates[expStateIdx]
			expState.CheckState(t, test.Name, &ctx)
		}
	}
}
