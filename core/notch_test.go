package core

type TestInstruction struct {
	Str   string
	Words []Word
}

// Test examples taken from dcpu-16_1.1.txt, and modified such that literal
// lengths indicate "next word" (4 hex digits) vs "embedded literal" (2 hex
// digits), and replacing labels with address values, and so that it works with
// the supported DCPU version.
var notchExample = []TestInstruction{
	// Try some basic stuff
	{"SET A, 0x0030", []Word{0x7c01, 0x0030}},
	{"SET [0x1000], 0x0020", []Word{0x7fc1, 0x0020, 0x1000}},
	{"SUB A, [0x1000]", []Word{0x7803, 0x1000}},
	{"IFN A, 16", []Word{0xc413}},
	{"SET PC, 0x001a", []Word{0x7f81, 0x001a}},
	// Do a loopy thing
	{"SET I, 10", []Word{0xacc1}},
	{"SET A, 0x2000", []Word{0x7c01, 0x2000}},
	{"SET [I+0x2000], [A]", []Word{0x22c1, 0x2000}},
	{"SUB I, 1", []Word{0x88c3}},
	{"IFN I, 0", []Word{0x84d3}},
	{"SET PC, 0x000d", []Word{0x7f81, 0x000d}},
	// Call a subroutine
	{"SET X, 4", []Word{0x9461}},
	{"JSR 0x0018", []Word{0x7c20, 0x0018}},
	{"SET PC, 0x001a", []Word{0x7f81, 0x001a}},
	{"SHL X, 4", []Word{0x946f}},
	{"SET PC, POP", []Word{0x6381}},
	// Hang forever. X should now be 0x40 if everything went right.
	{"SET PC, 0x001a", []Word{0x7f81, 0x001a}},
}

func concatTestInstructions(instructions []TestInstruction) []Word {
	var result []Word
	for _, ins := range instructions {
		result = append(result, ins.Words...)
	}
	return result
}
