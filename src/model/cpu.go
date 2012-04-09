package model

type CPUState struct {
	registers [8]Word // registers A-J, index with RegisterId.
	pc        Word    // Program counter.
	sp        Word    // Stack pointer.
	o         Word    // Overflow.
}

func (cpu *CPUState) Init() {
	for i := range cpu.registers {
		cpu.registers[i] = 0x0000
	}

	cpu.pc = 0x0000
	cpu.sp = 0xffff
	cpu.o = 0x0000
}

func (cpu *CPUState) Register(id RegisterId) Word {
	return cpu.registers[id]
}

func (cpu *CPUState) WriteRegister(id RegisterId, value Word) {
	cpu.registers[id] = value
}

func (cpu *CPUState) PC() Word {
	return cpu.pc
}

func (cpu *CPUState) WritePC(value Word) {
	cpu.pc = value
}

// Reads PC, and increments it (PC++).
func (cpu *CPUState) ReadIncPC() Word {
	value := cpu.pc
	cpu.pc++
	return value
}

func (cpu *CPUState) SP() Word {
	return cpu.sp
}

func (cpu *CPUState) WriteSP(value Word) {
	cpu.sp = value
}

// Reads SP, and increments it (SP++).
func (cpu *CPUState) ReadIncSP() Word {
	value := cpu.sp
	cpu.sp++
	return value
}

// Decrements SP and reads it (--SP).
func (cpu *CPUState) DecReadSP() Word {
	cpu.sp--
	return cpu.sp
}

func (cpu *CPUState) O() Word {
	return cpu.o
}

func (cpu *CPUState) WriteO(value Word) {
	cpu.o = value
}
