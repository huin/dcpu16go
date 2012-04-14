package model

type BasicCPU struct {
	registers [8]Word // registers A-J, index with RegisterId.
	pc        Word    // Program counter.
	sp        Word    // Stack pointer.
	o         Word    // Overflow.
}

func (cpu *BasicCPU) Init() {
	for i := range cpu.registers {
		cpu.registers[i] = 0x0000
	}

	cpu.pc = 0x0000
	cpu.sp = 0xffff
	cpu.o = 0x0000
}

func (cpu *BasicCPU) Register(id RegisterId) Word {
	return cpu.registers[id]
}

func (cpu *BasicCPU) WriteRegister(id RegisterId, value Word) {
	cpu.registers[id] = value
}

func (cpu *BasicCPU) PC() Word {
	return cpu.pc
}

func (cpu *BasicCPU) WritePC(value Word) {
	cpu.pc = value
}

// Reads PC, and increments it (PC++).
func (cpu *BasicCPU) ReadIncPC() Word {
	value := cpu.pc
	cpu.pc++
	return value
}

func (cpu *BasicCPU) SP() Word {
	return cpu.sp
}

func (cpu *BasicCPU) WriteSP(value Word) {
	cpu.sp = value
}

// Reads SP, and increments it (SP++).
func (cpu *BasicCPU) ReadIncSP() Word {
	value := cpu.sp
	cpu.sp++
	return value
}

// Decrements SP and reads it (--SP).
func (cpu *BasicCPU) DecReadSP() Word {
	cpu.sp--
	return cpu.sp
}

func (cpu *BasicCPU) O() Word {
	return cpu.o
}

func (cpu *BasicCPU) WriteO(value Word) {
	cpu.o = value
}
