package irc

type Handler interface {
	In(input []byte)
}
