package handlers

// Null (debug logs)
type NullHandler struct{}

func (nh NullHandler) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (nh NullHandler) Close() error {
	return nil
}

func (nh NullHandler) Flush() {}
