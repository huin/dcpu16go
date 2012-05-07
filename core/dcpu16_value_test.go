package core

import (
	"fmt"
	"testing"
)

func TestNoValueString(t *testing.T) {
	type Test struct {
		Expected    string
		ValueWord   Word
		HasNextWord bool
		NextWord    Word
	}
	tests := []Test{
		{"A", 0x00, false, 0},
		{"J", 0x07, false, 0},
		{"[A]", 0x08, false, 0},
		{"[J]", 0x0f, false, 0},
		{"[0x4321+A]", 0x10, true, 0x4321},
		{"[0x1234+J]", 0x17, true, 0x1234},
		{"POP", 0x18, false, 0},
		{"PEEK", 0x19, false, 0},
		{"PUSH", 0x1a, false, 0},
		{"SP", 0x1b, false, 0},
		{"PC", 0x1c, false, 0},
		{"O", 0x1d, false, 0},
		{"[0xdead]", 0x1e, true, 0xdead},
		{"0xbeef", 0x1f, true, 0xbeef},
		{"0x00", 0x20, false, 0},
		{"0x1f", 0x3f, false, 0},
	}

	var valueSet D16ValueSet

	for _, test := range tests {
		wordLoader := &FakeWordLoader{t: t}
		if test.HasNextWord {
			wordLoader.Words = []Word{test.NextWord}
		}
		value, err := valueSet.Value(test.ValueWord)
		if err != nil {
			t.Errorf("Unexpected error for value code %02x: %v",
				test.ValueWord, err)
			continue
		}
		value.LoadExtraWords(wordLoader)
		if !wordLoader.exhausted() {
			t.Errorf("Value %v (%02x) did not consume next word",
				value, test.ValueWord)
		}
		str := value.String()
		if test.Expected != str {
			t.Errorf("Value %v (%02x) disagrees on string repr (expected %q, got %q)",
				value, test.ValueWord, test.Expected, str)
		}
	}
}

func TestValueWrite(t *testing.T) {
	type Test struct {
		Value     Value
		InitCPU   D16CPU
		WriteWord Word
		ExpStates []StateChecker
		ExpCPU    D16CPU
	}

	tests := []Test{
		{
			Value:     &RegisterAddressValue{Reg: RegA},
			InitCPU:   D16CPU{registers: [8]Word{0x1234}},
			WriteWord: 0x5678,
			ExpStates: []StateChecker{
				&ExpMem{0x1234, []Word{0x5678}},
			},
			ExpCPU: D16CPU{registers: [8]Word{0x1234}},
		},
		{
			Value:     &RegisterRelAddressValue{extraWord{5}, RegA},
			InitCPU:   D16CPU{registers: [8]Word{0x1234}},
			WriteWord: 0x5678,
			ExpStates: []StateChecker{
				&ExpMem{0x1239, []Word{0x5678}},
			},
			ExpCPU: D16CPU{registers: [8]Word{0x1234}},
		},
		{
			Value:     PopValue{},
			InitCPU:   D16CPU{sp: 0xfffe},
			WriteWord: 0x5678,
			ExpStates: []StateChecker{
				&ExpMem{0xfffe, []Word{0x5678, 0x0000}},
			},
			ExpCPU: D16CPU{sp: 0xffff},
		},
		{
			Value:     PeekValue{},
			InitCPU:   D16CPU{sp: 0xfffe},
			WriteWord: 0x5678,
			ExpStates: []StateChecker{
				&ExpMem{0xfffe, []Word{0x5678, 0x0000}},
			},
			ExpCPU: D16CPU{sp: 0xfffe},
		},
		{
			Value:     PushValue{},
			InitCPU:   D16CPU{sp: 0xffff},
			WriteWord: 0x5678,
			ExpStates: []StateChecker{
				&ExpMem{0xfffe, []Word{0x5678, 0x0000}},
			},
			ExpCPU: D16CPU{sp: 0xfffe},
		},
	}

	for _, test := range tests {
		var state D16MachineState
		state.Init()
		state.D16CPU = test.InitCPU
		test.Value.Write(&state, test.WriteWord)
		name := fmt.Sprintf("%v write 0x%04x", test.Value, test.WriteWord)
		for _, expState := range test.ExpStates {
			expState.StateCheck(t, name, &state)
		}
		if !CPUEquals(&test.ExpCPU, &state.D16CPU) {
			t.Errorf("%v\ngot CPU state %#v\n     expected %#v", test.Value, state.D16CPU, test.ExpCPU)
		}
	}
}

func TestValueRead(t *testing.T) {
	type Test struct {
		Value         Value
		InitCPU       D16CPU
		InitMemOffset Word
		InitMem       []Word
		ExpRead       Word
		ExpCPU        D16CPU
	}

	tests := []Test{
		{
			Value:         &RegisterAddressValue{Reg: RegA},
			InitCPU:       D16CPU{registers: [8]Word{0x1234}},
			InitMemOffset: 0x1234,
			InitMem:       []Word{0x5678},
			ExpRead:       0x5678,
			ExpCPU:        D16CPU{registers: [8]Word{0x1234}},
		},
		{
			Value:         &RegisterRelAddressValue{extraWord{5}, RegA},
			InitCPU:       D16CPU{registers: [8]Word{0x1234}},
			InitMemOffset: 0x1239,
			InitMem:       []Word{0x5678},
			ExpRead:       0x5678,
			ExpCPU:        D16CPU{registers: [8]Word{0x1234}},
		},
		{
			Value:         PopValue{},
			InitCPU:       D16CPU{sp: 0xfffe},
			InitMemOffset: 0xfffe,
			InitMem:       []Word{0x5678, 0x0000},
			ExpRead:       0x5678,
			ExpCPU:        D16CPU{sp: 0xffff},
		},
		{
			Value:         PeekValue{},
			InitCPU:       D16CPU{sp: 0xfffe},
			InitMemOffset: 0xfffe,
			InitMem:       []Word{0x5678, 0x0000},
			ExpRead:       0x5678,
			ExpCPU:        D16CPU{sp: 0xfffe},
		},
		{
			Value:         PushValue{},
			InitCPU:       D16CPU{sp: 0xffff},
			InitMemOffset: 0xfffe,
			InitMem:       []Word{0x5678, 0x0000},
			ExpRead:       0x5678,
			ExpCPU:        D16CPU{sp: 0xfffe},
		},
	}

	for _, test := range tests {
		var state D16MachineState
		state.Init()
		copy(state.D16MemoryState.Data[test.InitMemOffset:], test.InitMem)
		state.D16CPU = test.InitCPU
		result := test.Value.Read(&state)
		if test.ExpRead != result {
			t.Errorf("%v returned 0x%04x, expected 0x%04x", test.Value, result, test.ExpRead)
		}
		if !CPUEquals(&test.ExpCPU, &state.D16CPU) {
			t.Errorf("%v\ngot CPU state %#v\n     expected %#v", test.Value, state.D16CPU, test.ExpCPU)
		}
	}
}
