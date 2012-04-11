package model

func Step(ctx Context) error {
	op, err := OperationLoad(ctx)
	if err != nil {
		return err
	}
	return op.Execute(ctx)
}
