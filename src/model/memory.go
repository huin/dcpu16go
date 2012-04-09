package model

const MemorySize = 0x10000

type MemoryState struct {
	Data [MemorySize]Word
}

func (mem *MemoryState) ReadMemory(address Word) Word {
	return mem.Data[address]
}

func (mem *MemoryState) WriteMemory(address Word, value Word) {
	mem.Data[address] = value
}
