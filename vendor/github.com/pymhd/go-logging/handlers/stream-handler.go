package handlers

import (
	"os"
)

//Console
type StreamHandler struct{}

func (sh StreamHandler) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (sh StreamHandler) Close() error {
	return nil
}

func (sh StreamHandler) Flush() {
}
