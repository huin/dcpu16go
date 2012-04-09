package model

func Step(ctx Context) error {
	op, err := OperationLoad(ctx)
	if err != nil {
		return err
	}
	op.Execute(ctx)
	return nil
}
