package core

const MemorySize = 0x10000

type D16MemoryState struct {
	Data [MemorySize]Word
}

func (mem *D16MemoryState) ReadMemory(address Word) Word {
	return mem.Data[address]
}

func (mem *D16MemoryState) WriteMemory(address Word, value Word) {
	mem.Data[address] = value
}
