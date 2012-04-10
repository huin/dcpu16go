package model

import (
	"testing"
)

type ExpMem struct {
	Offset   int
	Expected []Word
}

func (exp *ExpMem) CheckMem(t *testing.T, testName string, mem *MemoryState) {
	for i := range exp.Expected {
		if exp.Expected[i] != mem.Data[exp.Offset+i] {
			t.Errorf("%s: memory state at 0x%04x", testName, exp.Offset+i)
			t.Errorf("expected: %#v", exp.Expected[i])
			t.Errorf("got:      %#v", mem.Data[exp.Offset+i])
			break
		}
	}
}

func TestStep(t *testing.T) {
	type Test struct {
		Name     string
		NumSteps int
		InitMem  []Word
		ExpCPU   CPUState
		ExpMems  []ExpMem
	}

	tests := []Test{
		{
			Name:     "SET A, 0x0030",
			NumSteps: 1,
			InitMem:  []Word{0x7c01, 0x0030},
			ExpCPU:   CPUState{registers: [8]Word{0x0030, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0002, sp: 0xffff},
			ExpMems:  []ExpMem{{0x0001, []Word{0x0030}}},
		},
		{
			Name:     "SET [0x0003], 0xbeef",
			NumSteps: 1,
			InitMem:  []Word{0x7de1, 0x0003, 0xbeef, 0x0000},
			ExpCPU:   CPUState{registers: [8]Word{0, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0003, sp: 0xffff},
			ExpMems:  []ExpMem{{0x0003, []Word{0xbeef}}},
		},
		{
			Name:     "SET A, 0x0030 / SET PUSH, A / SET C, PEEK / SET B, POP",
			NumSteps: 4,
			InitMem: []Word{
				0x7c01, 0x0030, // SET A, 0x0030
				0x01a1, // SET PUSH, A
				0x6421, // SET C, PEEK
				0x6011, // SET B, POP
			},
			ExpCPU:  CPUState{registers: [8]Word{0x0030, 0x0030, 0x0030, 0, 0, 0, 0, 0}, pc: 0x0005, sp: 0xffff},
			ExpMems: []ExpMem{{0xfffe, []Word{0x0030, 0x0000}}},
		},
	}

	for i := range tests {
		test := &tests[i]

		ctx := StandardContext{}
		ctx.CPUState.Init()
		copy(ctx.MemoryState.Data[0:], test.InitMem)

		for i := 0; i < test.NumSteps; i++ {
			Step(&ctx)
		}

		if !CPUEquals(&test.ExpCPU, &ctx.CPUState) {
			t.Errorf("%s: CPU state", test.Name)
			t.Errorf("expected: %#v", test.ExpCPU)
			t.Errorf("got:      %#v", ctx.CPUState)
		}

		for i := range test.ExpMems {
			test.ExpMems[i].CheckMem(t, test.Name, &ctx.MemoryState)
		}
	}
}
