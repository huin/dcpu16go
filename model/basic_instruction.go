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

	// Use separate pools of values so that two values in play at the same time
	// don't interfere (see doc for BasicValueSet).
	aValueSet BasicValueSet
	bValueSet BasicValueSet
}

func (is *BasicInstructionSet) init() {
	is.binarySet = [0x10]BinaryInstruction{
		nil, // Indicates unary instruction.
		&is.setInst, &is.addInst, &is.subInst, &is.mulInst,
		&is.divInst, &is.modInst, &is.shlInst, &is.shrInst,
		&is.andInst, &is.borInst, &is.xorInst, &is.ifeInst,
		&is.ifnInst, &is.ifgInst, &is.ifbInst,
	}
}

func (is *BasicInstructionSet) unaryInstruction(upper6, middle6 Word) (instruction UnaryInstruction, a Value, err error) {
	if middle6 != 0x001 {
		err = InvalidUnaryOpCodeError(middle6)
		return
	}
	instruction = &is.jsrInst
	a, err = is.aValueSet.Value(upper6)
	return
}

func (is *BasicInstructionSet) binaryInstruction(upper6, middle6, lower4 Word) (instruction BinaryInstruction, a, b Value, err error) {
	opCode := lower4
	if int(opCode) >= len(is.binarySet) {
		// Shouldn't happen with the given set.
		err = InvalidBinaryOpCodeError(opCode)
		return
	}
	instruction = is.binarySet[opCode]
	if instruction == nil {
		// Shouldn't happen with the given set.
		err = InvalidBinaryOpCodeError(opCode)
		return
	}
	a, err = is.aValueSet.Value(middle6)
	if err != nil {
		return
	}
	b, err = is.bValueSet.Value(upper6)
	if err != nil {
		return
	}
	return
}

func (is *BasicInstructionSet) NumExtraWords(w Word) (Word, error) {
	if !is.initialized {
		is.init()
	}

	upper6, middle6, lower4 := splitOpWord(w)

	if lower4 != 0 {
		// Binary instruction.
		_, a, b, err := is.binaryInstruction(upper6, middle6, lower4)
		if err != nil {
			return 0, err
		}
		return a.NumExtraWords() + b.NumExtraWords(), nil
	}

	// Unary instruction.
	_, a, err := is.unaryInstruction(upper6, middle6)
	if err != nil {
		return 0, err
	}
	return a.NumExtraWords(), nil
}

func (is *BasicInstructionSet) Instruction(w Word) (Instruction, error) {
	if !is.initialized {
		is.init()
	}

	upper6, middle6, lower4 := splitOpWord(w)

	if lower4 != 0 {
		// Binary instruction.
		instruction, a, b, err := is.binaryInstruction(upper6, middle6, lower4)
		if err != nil {
			return nil, err
		}
		instruction.SetBinaryValue(a, b)
		return instruction, nil
	}

	// Unary instruction.
	instruction, a, err := is.unaryInstruction(upper6, middle6)
	if err != nil {
		return nil, err
	}
	instruction.SetUnaryValue(a)
	return instruction, nil
}
