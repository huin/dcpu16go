package core

type (
	Word   uint16
	SWord  int16
	DWord  uint32
	DSWord int32
)

func Signed(w SWord) Word {
	return Word(w)
}

func (w Word) AsDSigned() DSWord {
	return DSWord(SWord(w))
}

func (dw DWord) Split() (Word, Word) {
	return Word(dw >> 16), Word(dw & 0xffff)
}

func (dsw DSWord) Split() (Word, Word) {
	return DWord(dsw).Split()
}
