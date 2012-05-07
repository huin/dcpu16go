package core

import (
	"fmt"
	"testing"
)

func TestValueString(t *testing.T) {
	type Test struct {
		Expected    string
		ValueWord   Word
		HasNextWord bool
		NextWord    Word
		AsValueB    bool
	}
	tests := []Test{
		{"A", 0x00, false, 0, false},
		{"J", 0x07, false, 0, false},
		{"[A]", 0x08, false, 0, false},
		{"[J]", 0x0f, false, 0, false},
		{"[A+0x4321]", 0x10, true, 0x4321, false},
		{"[J+0x1234]", 0x17, true, 0x1234, false},
		{"POP", 0x18, false, 0, false},
		{"PUSH", 0x18, false, 0, true},
		{"PEEK", 0x19, false, 0, false},
		{"PICK 0xcafe", 0x1a, true, 0xcafe, false},
		{"SP", 0x1b, false, 0, false},
		{"PC", 0x1c, false, 0, false},
		{"EX", 0x1d, false, 0, false},
		{"[0xdead]", 0x1e, true, 0xdead, false},
		{"0xbeef", 0x1f, true, 0xbeef, false},
		{"-1", 0x20, false, 0, false},
		{"0", 0x21, false, 0, false},
		{"1", 0x22, false, 0, false},
		{"30", 0x3f, false, 0, false},
	}

	var valueSet D16ValueSet

	for _, test := range tests {
		wordLoader := &FakeWordLoader{t: t}
		if test.HasNextWord {
			wordLoader.Words = []Word{test.NextWord}
		}
		value, err := valueSet.Value(test.ValueWord, test.AsValueB)
		if err != nil {
			t.Errorf("Unexpected error for value code %02x: %v",
				test.ValueWord, err)
			continue
		}
		value.LoadExtraWords(wordLoader)
		if !wordLoader.exhausted() {
			t.Errorf("Value %#v (%02x) did not consume next word",
				value, test.ValueWord)
		}
		str := value.String()
		if test.Expected != str {
			t.Errorf("Value %#v (%02x) disagrees on ValueString(%t) repr (expected %q, got %q)",
				value, test.ValueWord, test.AsValueB, test.Expected, str)
		}
	}
}

func TestErrorValue(t *testing.T) {
	var valueSet D16ValueSet

	// In-instruction literals cannot be used as value B.
	// (Presumably next-word literals can be, but uselessly).
	value, err := valueSet.Value(0x20, true)
	if err == nil {
		t.Errorf("Expected error for value code 0x20, but got value: %#v", value)
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
		{
			Value:     &PickValue{extraWord{0x0001}},
			InitCPU:   D16CPU{sp: 0xfffe},
			WriteWord: 0x1234,
			ExpStates: []StateChecker{
				&ExpMem{0xffff, []Word{0x1234}},
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
			Value:         &PickValue{extraWord{0x0001}},
			InitCPU:       D16CPU{sp: 0xfffe},
			InitMemOffset: 0xffff,
			InitMem:       []Word{0x1234},
			ExpRead:       0x1234,
			ExpCPU:        D16CPU{sp: 0xfffe},
		},
		{
			// Literal -1
			Value:   &LiteralValue{Literal: 0xffff},
			InitCPU: D16CPU{sp: 0xffff},
			ExpRead: 0xffff,
			ExpCPU:  D16CPU{sp: 0xffff},
		},
		{
			Value:   &LiteralValue{Literal: 30},
			InitCPU: D16CPU{sp: 0xffff},
			ExpRead: 30,
			ExpCPU:  D16CPU{sp: 0xffff},
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
