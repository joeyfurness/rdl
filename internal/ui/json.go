package ui

import (
	"encoding/json"
	"io"
)

// JSONEmitter writes events as newline-delimited JSON (NDJSON).
type JSONEmitter struct {
	w   io.Writer
	enc *json.Encoder
}

// NewJSONEmitter creates a JSONEmitter that writes to w.
func NewJSONEmitter(w io.Writer) *JSONEmitter {
	return &JSONEmitter{
		w:   w,
		enc: json.NewEncoder(w),
	}
}

// Emit encodes the event as a single JSON line.
func (e *JSONEmitter) Emit(ev Event) {
	_ = e.enc.Encode(ev)
}
