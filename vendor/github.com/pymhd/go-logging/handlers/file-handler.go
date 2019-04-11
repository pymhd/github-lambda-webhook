package handlers

import (
	"os"
)

//File
type FileHandler struct {
	fd *os.File
}

func createFile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	f.Close()
}

func (fh FileHandler) Write(p []byte) (n int, err error) {
	return fh.fd.Write(p)
}

func (fh FileHandler) Close() error {
	return fh.fd.Close()
}

func (fh FileHandler) Flush() {
	fh.fd.Sync()
}

func NewFileHandler(filename string) FileHandler {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		createFile(filename)
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	h := FileHandler{fd: f}
	return h
}
