package model

const MemorySize = 0x10000

type BasicMemoryState struct {
	Data [MemorySize]Word
}

func (mem *BasicMemoryState) ReadMemory(address Word) Word {
	return mem.Data[address]
}

func (mem *BasicMemoryState) WriteMemory(address Word, value Word) {
	mem.Data[address] = value
}
