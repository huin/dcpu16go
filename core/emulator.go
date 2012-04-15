package model

func Step(state MachineState) error {
	instruction, err := InstructionLoad(state, state)
	if err != nil {
		return err
	}
	return instruction.Execute(state)
}
