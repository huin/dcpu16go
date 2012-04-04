package model

type WordLoader interface {
	WordLoad() Word
}

type Context interface {
	WordLoader
}
