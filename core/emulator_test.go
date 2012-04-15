package core

import (
	"testing"
)

type ExpMem struct {
	Offset   int
	Expected []Word
}

func (exp *ExpMem) StateCheck(t *testing.T, testName string, mem *D16MachineState) bool {
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

type StateChecker interface {
	StateCheck(t *testing.T, testName string, state *D16MachineState) bool
}

type ErrState struct {
	Name     string
	NumSteps int
}

func (exp *ErrState) StateCheck(t *testing.T, testName string, state *D16MachineState) bool {
	testName = testName + "::" + exp.Name
	for i := 0; i < (exp.NumSteps - 1); i++ {
		err := Step(state)
		if err != nil {
			t.Fatalf("%s: unexpected emulation error: %v", testName, err)
		}
	}
	err := Step(state)
	if err == nil {
		t.Fatalf("%s: no expected emulation error", testName)
	}
	return true
}

type ExpState struct {
	Name     string
	NumSteps int
	CPU      D16CPU
	Mems     []ExpMem
}

func (exp *ExpState) StateCheck(t *testing.T, testName string, state *D16MachineState) bool {
	testName = testName + "::" + exp.Name

	for i := 0; i < exp.NumSteps; i++ {
		err := Step(state)
		if err != nil {
			t.Fatalf("%s: unexpected emulation error: %v", testName, err)
		}
	}

	if !CPUEquals(&exp.CPU, &state.D16CPU) {
		t.Errorf("%s: CPU state", testName)
		t.Errorf("expected: %#v", exp.CPU)
		t.Errorf("got:      %#v", state.D16CPU)
		return false
	}

	for i := range exp.Mems {
		if !exp.Mems[i].StateCheck(t, testName, state) {
			return false
		}
	}

	return true
}

func TestStep(t *testing.T) {
	type Test struct {
		Name      string
		InitMem   []Word
		ExpStates []StateChecker
	}

	tests := []Test{
		{
			Name:    "SET A, 0x0030",
			InitMem: []Word{0x7c01, 0x0030},
			ExpStates: []StateChecker{
				&ExpState{
					NumSteps: 1,
					CPU:      D16CPU{registers: [8]Word{0x0030, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0002, sp: 0xffff},
					Mems:     []ExpMem{{0x0001, []Word{0x0030}}},
				},
			},
		},
		{
			Name:    "SET [0x0003], 0xbeef",
			InitMem: []Word{0x7de1, 0x0003, 0xbeef, 0x0000},
			ExpStates: []StateChecker{
				&ExpState{
					NumSteps: 1,
					CPU:      D16CPU{registers: [8]Word{0, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0003, sp: 0xffff},
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
			ExpStates: []StateChecker{
				&ExpState{
					NumSteps: 4,
					CPU:      D16CPU{registers: [8]Word{0x0030, 0x0030, 0x0030, 0, 0, 0, 0, 0}, pc: 0x0005, sp: 0xffff},
					Mems:     []ExpMem{{0xfffe, []Word{0x0030, 0x0000}}},
				},
			},
		},
		{
			Name:    "unknown instruction",
			InitMem: []Word{0xfff0},
			ExpStates: []StateChecker{
				&ErrState{NumSteps: 1},
			},
		},
		{
			Name:    "IFE 0, 1 / unknown instruction",
			InitMem: []Word{0x860c, 0xfff0},
			ExpStates: []StateChecker{
				&ErrState{NumSteps: 1},
				&ExpState{
					Name:     "state check",
					NumSteps: 0,
					CPU:      D16CPU{pc: 0x0002, sp: 0xffff},
				},
			},
		},
		{
			Name: "Notch example",
			InitMem: []Word{
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
			},
			ExpStates: []StateChecker{
				&ExpState{
					Name:     "after basic stuff",
					NumSteps: 4,
					CPU:      D16CPU{registers: [8]Word{0x0010}, pc: 0x000a, sp: 0xffff, o: 0x0000},
					Mems: []ExpMem{
						{0x1000, []Word{0x0020}},
					},
				},
				&ExpState{
					Name:     "after loop init",
					NumSteps: 2,
					CPU:      D16CPU{registers: [8]Word{0x2000, 0, 0, 0, 0, 0, 10, 0}, pc: 0x000d, sp: 0xffff, o: 0x0000},
				},
				&ExpState{
					Name:     "after first test+iteration",
					NumSteps: 4,
					CPU:      D16CPU{registers: [8]Word{0x2000, 0, 0, 0, 0, 0, 9, 0}, pc: 0x000d, sp: 0xffff, o: 0x0000},
					Mems: []ExpMem{
						{0x2010, []Word{0x0000}},
					},
				},
				&ExpState{
					Name:     "after loop completion",
					NumSteps: 4*8 + 3,
					CPU:      D16CPU{registers: [8]Word{0x2000, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0013, sp: 0xffff, o: 0x0000},
					Mems: []ExpMem{
						{0x2000, []Word{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000}},
					},
				},
				&ExpState{
					Name:     "before JSR",
					NumSteps: 1,
					CPU:      D16CPU{registers: [8]Word{0x2000, 0, 0, 0x0004}, pc: 0x0014, sp: 0xffff, o: 0x0000},
				},
				&ExpState{
					Name:     "in testsub",
					NumSteps: 1,
					CPU:      D16CPU{registers: [8]Word{0x2000, 0, 0, 0x0004}, pc: 0x0018, sp: 0xfffe, o: 0x0000},
					Mems: []ExpMem{
						{0xfffe, []Word{0x0016, 0x0000}},
					},
				},
				&ExpState{
					Name:     "returned from testsub",
					NumSteps: 2,
					CPU:      D16CPU{registers: [8]Word{0x2000, 0, 0, 0x0040}, pc: 0x0016, sp: 0xffff, o: 0x0000},
				},
				&ExpState{
					Name:     "entering crash loop",
					NumSteps: 1,
					CPU:      D16CPU{registers: [8]Word{0x2000, 0, 0, 0x0040}, pc: 0x001a, sp: 0xffff, o: 0x0000},
				},
				&ExpState{
					Name:     "still in crash loop",
					NumSteps: 10,
					CPU:      D16CPU{registers: [8]Word{0x2000, 0, 0, 0x0040}, pc: 0x001a, sp: 0xffff, o: 0x0000},
				},
			},
		},
	}

	for i := range tests {
		test := &tests[i]

		state := D16MachineState{}
		state.D16CPU.Init()
		copy(state.D16MemoryState.Data[0:], test.InitMem)

		for _, expState := range test.ExpStates {
			expState.StateCheck(t, test.Name, &state)
		}
	}
}
