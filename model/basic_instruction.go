package model

func splitOpWord(w Word) (upper6, middle6, lower4 Word) {
	upper6 = (w >> 10) & 0x003f
	middle6 = (w >> 4) & 0x003f
	lower4 = w & 0x000f
	return
}

// BasicInstructionSet is the table of basic instructions supported by the
// DCPU-16. This implementation avoids memory allocation, but the instruction
// returned is only good for use until the next call to the Instruction method.
type BasicInstructionSet struct {
	initialized bool
	unarySet    [2]UnaryInstruction
	jsrInst     JsrInst
	binarySet   [0x10]BinaryInstruction
	setInst     SetInst
	addInst     AddInst
	subInst     SubInst
	mulInst     MulInst
	divInst     DivInst
	modInst     ModInst
	shlInst     ShlInst
	shrInst     ShrInst
	andInst     AndInst
	borInst     BorInst
	xorInst     XorInst
	ifeInst     IfeInst
	ifnInst     IfnInst
	ifgInst     IfgInst
	ifbInst     IfbInst
}

func (is *BasicInstructionSet) init() {
	is.unarySet = [2]UnaryInstruction{
		nil, // Reserved for future expansion.
		&is.jsrInst,
	}
	is.binarySet = [0x10]BinaryInstruction{
		nil, // Indicates unary instruction.
		&is.setInst, &is.addInst, &is.subInst, &is.mulInst,
		&is.divInst, &is.modInst, &is.shlInst, &is.shrInst,
		&is.andInst, &is.borInst, &is.xorInst, &is.ifeInst,
		&is.ifnInst, &is.ifgInst, &is.ifbInst,
	}
}

func (is *BasicInstructionSet) NumExtraWords(w Word) (Word, error) {
	if !is.initialized {
		is.init()
	}

	upper6, middle6, opCode := splitOpWord(w)

	if opCode != 0 {
		// Binary instruction.
		if int(opCode) >= len(is.binarySet) {
			// Shouldn't happen with the given set.
			return 0, InvalidBinaryOpCodeError(opCode)
		}
		instruction := is.binarySet[opCode]
		if instruction == nil {
			// Shouldn't happen with the given set.
			return 0, InvalidBinaryOpCodeError(opCode)
		}
		a := ValueFromWord(middle6)
		b := ValueFromWord(upper6)
		return a.NumExtraWords() + b.NumExtraWords(), nil
	}

	// Unary instruction.
	uniOpCode := middle6
	if int(uniOpCode) >= len(is.unarySet) {
		return 0, InvalidUnaryOpCodeError(uniOpCode)
	}
	instruction := is.unarySet[uniOpCode]
	if instruction == nil {
		return 0, InvalidUnaryOpCodeError(uniOpCode)
	}
	a := ValueFromWord(upper6)
	return a.NumExtraWords(), nil
}

func (is *BasicInstructionSet) Instruction(w Word) (Instruction, error) {
	if !is.initialized {
		is.init()
	}

	upper6, middle6, opCode := splitOpWord(w)

	if opCode != 0 {
		// Binary instruction.
		if int(opCode) >= len(is.binarySet) {
			// Shouldn't happen with the given set.
			return nil, InvalidBinaryOpCodeError(opCode)
		}
		instruction := is.binarySet[opCode]
		if instruction == nil {
			// Shouldn't happen with the given set.
			return nil, InvalidBinaryOpCodeError(opCode)
		}
		a := ValueFromWord(middle6)
		b := ValueFromWord(upper6)
		instruction.SetBinaryValue(a, b)
		return instruction, nil
	}

	// Unary instruction.
	uniOpCode := middle6
	if int(uniOpCode) >= len(is.unarySet) {
		return nil, InvalidUnaryOpCodeError(uniOpCode)
	}
	instruction := is.unarySet[uniOpCode]
	if instruction == nil {
		return nil, InvalidUnaryOpCodeError(uniOpCode)
	}
	a := ValueFromWord(upper6)
	instruction.SetUnaryValue(a)
	return instruction, nil
}
