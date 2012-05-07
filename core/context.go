package core

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
	EX() Word
	WriteEX(value Word)
}

func CPUEquals(a, b CPU) bool {
	for id := RegA; id <= RegJ; id++ {
		if a.Register(id) != b.Register(id) {
			return false
		}
	}
	return a.PC() == b.PC() && a.SP() == b.SP() && a.EX() == b.EX()
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

type D16MachineState struct {
	D16InstructionSet
	D16CPU
	D16MemoryState
}

func (state *D16MachineState) Init() {
	state.D16CPU.Init()
}

func (state *D16MachineState) WordLoad() (Word, error) {
	return state.D16MemoryState.ReadMemory(state.ReadIncPC()), nil
}

func (state *D16MachineState) SkipWords(count Word) error {
	state.WritePC(state.PC() + count)
	return nil
}
