package ui

import (
	"io"
)

// The UI is a basic UI that reads and writes from a standard Go reader
// and writer. It is safe to be called from multiple goroutines. Machine
// readable output is simply logged for this UI.
type UI struct {
	Reader      io.Reader
	Writer      io.Writer
	ErrorWriter io.Writer
}
