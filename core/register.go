package core

type RegisterId uint8

const (
	RegA RegisterId = iota
	RegB
	RegC
	RegX
	RegY
	RegZ
	RegI
	RegJ
)

func (r RegisterId) String() string {
	switch r {
	case RegA:
		return "A"
	case RegB:
		return "B"
	case RegC:
		return "C"
	case RegX:
		return "X"
	case RegY:
		return "Y"
	case RegZ:
		return "Z"
	case RegI:
		return "I"
	case RegJ:
		return "J"
	}
	panic(r)
}
