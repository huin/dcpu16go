package model

import (
	"testing"
)

func TestStep(t *testing.T) {
	type Test struct {
		Name     string
		NumSteps int
		InitMem  []Word
		ExpCPU   CPUState
		ExpMem   []Word
	}

	tests := []Test{
		{
			Name:     "SET A, 0x0030",
			NumSteps: 1,
			InitMem:  []Word{0x7c01, 0x0030}, // SET A, 0x0030
			ExpCPU:   CPUState{registers: [8]Word{0x0030, 0, 0, 0, 0, 0, 0, 0}, pc: 0x0002, sp: 0xffff},
			ExpMem:   []Word{0x7c01, 0x0030},
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
	}
}
