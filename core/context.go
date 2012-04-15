package model

type WordLoader interface {
	WordLoad() (Word, error)
	SkipWords(Word) error
}

type CPU interface {
	Register(id RegisterId) Word
	WriteRegister(id RegisterId, value Word)
	PC() Word
	WritePC(value Word)
	// Reads PC, and increments it (PC++).
	ReadIncPC() Word
	SP() Word
	WriteSP(value Word)
	// Reads SP, and increments it (SP++).
	ReadIncSP() Word
	// Decrements SP and reads it (--SP).
	DecReadSP() Word
	O() Word
	WriteO(value Word)
}

func CPUEquals(a, b CPU) bool {
	for id := RegA; id <= RegJ; id++ {
		if a.Register(id) != b.Register(id) {
			return false
		}
	}
	return a.PC() == b.PC() && a.SP() == b.SP() && a.O() == b.O()
}

type Memory interface {
	ReadMemory(address Word) Word
	WriteMemory(address Word, value Word)
}

type MachineState interface {
	WordLoader
	InstructionSet
	CPU
	Memory
}

type BasicMachineState struct {
	BasicInstructionSet
	BasicCPU
	BasicMemoryState
}

func (state *BasicMachineState) Init() {
	state.BasicCPU.Init()
}

func (state *BasicMachineState) WordLoad() (Word, error) {
	return state.BasicMemoryState.ReadMemory(state.ReadIncPC()), nil
}

func (state *BasicMachineState) SkipWords(count Word) error {
	state.WritePC(state.PC() + count)
	return nil
}
