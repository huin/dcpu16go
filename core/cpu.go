package model

type D16CPU struct {
	registers [8]Word // registers A-J, index with RegisterId.
	pc        Word    // Program counter.
	sp        Word    // Stack pointer.
	o         Word    // Overflow.
}

func (cpu *D16CPU) Init() {
	for i := range cpu.registers {
		cpu.registers[i] = 0x0000
	}

	cpu.pc = 0x0000
	cpu.sp = 0xffff
	cpu.o = 0x0000
}

func (cpu *D16CPU) Register(id RegisterId) Word {
	return cpu.registers[id]
}

func (cpu *D16CPU) WriteRegister(id RegisterId, value Word) {
	cpu.registers[id] = value
}

func (cpu *D16CPU) PC() Word {
	return cpu.pc
}

func (cpu *D16CPU) WritePC(value Word) {
	cpu.pc = value
}

// Reads PC, and increments it (PC++).
func (cpu *D16CPU) ReadIncPC() Word {
	value := cpu.pc
	cpu.pc++
	return value
}

func (cpu *D16CPU) SP() Word {
	return cpu.sp
}

func (cpu *D16CPU) WriteSP(value Word) {
	cpu.sp = value
}

// Reads SP, and increments it (SP++).
func (cpu *D16CPU) ReadIncSP() Word {
	value := cpu.sp
	cpu.sp++
	return value
}

// Decrements SP and reads it (--SP).
func (cpu *D16CPU) DecReadSP() Word {
	cpu.sp--
	return cpu.sp
}

func (cpu *D16CPU) O() Word {
	return cpu.o
}

func (cpu *D16CPU) WriteO(value Word) {
	cpu.o = value
}
