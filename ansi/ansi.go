// Package ansi supports writing ANSI escape sequences.
//
// See https://en.wikipedia.org/wiki/ANSI_escape_code for a general overview of
// ANSI codes.
package ansi

import (
	"bytes"
	"io"

	"golang.org/x/term"
)

// A Coder wraps an io.Writer and adds features for emitting ANSI escape
// sequences.  An Coder is also an io.Writer that delegates writes to its
// underlying writer.
type Coder struct {
	w      io.Writer
	isTerm bool
}

// Write satisfies io.Writer by delegating to the wrapped writer.
func (c Coder) Write(data []byte) (int, error) { return c.w.Write(data) }

// Esc emits an escape sequence to w having the form:
//
//    ESC start [data]
//
// It returns the number of bytes written and any error from w.Write.
// If start == 0 this is equivalent to io.WriteString(w, data).
func Esc(w io.Writer, start byte, data string) (int, error) {
	if start == 0 {
		return io.WriteString(w, data)
	}
	size := 2 + len(data) // 2 for ESC start
	buf := make([]byte, size)
	buf[0] = '\033' // ESC
	buf[1] = start
	copy(buf[2:], data)
	return w.Write(buf)
}

// Set returns a writer that unconditionally wraps each write with the
// specified escape sequence having the form:
//
//    ESC start [prefix] <data> [suffix] [ESC end]
//
// If prefix == "" the prefix is omitted.
// If suffix == "" the suffix is omitted.
// If end == 0 the trailer is omitted.
func (c Coder) Set(start byte, prefix string, end byte, suffix string) io.Writer {
	return escWriter{
		start:  start,
		end:    end,
		prefix: prefix,
		suffix: suffix,
		w:      c.w,
	}
}

// SetIf returns a writer that wraps each write with the specified escape
// sequence if the underlying writer is attached to a terminal. Otherwise,
// writes are not wrapped. The semantics of start, end, and suffix are as
// described for Set.
func (c Coder) SetIf(start byte, prefix string, end byte, suffix string) io.Writer {
	if c.isTerm {
		return c.Set(start, prefix, end, suffix)
	}
	return c.w
}

// NewCoder constructs a new coder that writes to w.
func NewCoder(w io.Writer) *Coder {
	env := &Coder{w: w}
	if f, ok := w.(interface {
		Fd() uintptr
	}); ok {
		env.isTerm = term.IsTerminal(int(f.Fd()))
	}
	return env
}

type escWriter struct {
	start, end     byte
	prefix, suffix string
	w              io.Writer
}

func (e escWriter) Write(data []byte) (int, error) {
	size := len(data) + len(e.prefix) + len(e.suffix) + 2 + 2 // 2 for ESC start, 2 for ESC end
	buf := bytes.NewBuffer(make([]byte, 0, size))
	Esc(buf, e.start, e.prefix)
	buf.Write(data)
	Esc(buf, e.end, e.suffix)
	return e.w.Write(buf.Bytes())
}
