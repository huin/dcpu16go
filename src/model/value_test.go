package model

import (
	"testing"
)

func TestNoValueString(t *testing.T) {
	type T struct {
		Expected    string
		ValueWord   Word
		HasNextWord bool
		NextWord    Word
	}
	tests := []T{
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

	for _, test := range tests {
		wordLoader := &FakeWordLoader{t: t}
		if test.HasNextWord {
			wordLoader.Words = []Word{test.NextWord}
		}
		value := ValueFromWord(test.ValueWord)
		value.LoadOpValue(wordLoader)
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
