package model

func Step(state MachineState) error {
	op, err := OperationLoad(state)
	if err != nil {
		return err
	}
	return op.Execute(state)
}
