package handlers

type Handler interface {
	Write([]byte) (int, error)
	Close() error
	Flush()
}
