package ui

import (
	"fmt"
	"io"
)

// QuietEmitter only prints error events to stderr.
type QuietEmitter struct {
	w io.Writer
}

// NewQuietEmitter creates a QuietEmitter that writes to w (typically os.Stderr).
func NewQuietEmitter(w io.Writer) *QuietEmitter {
	return &QuietEmitter{w: w}
}

// Emit writes the event only if it is an error event.
func (q *QuietEmitter) Emit(ev Event) {
	if ev.Type == "error" {
		fmt.Fprintf(q.w, "error: %s\n", ev.Message)
	}
}
