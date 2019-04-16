package ansi

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

// Verify that directly writing to a Coder delegates correctly.
func TestWrite(t *testing.T) {
	const input = "OK"

	var buf bytes.Buffer
	c := NewCoder(&buf)

	io.WriteString(c, input)
	if got := buf.String(); got != input {
		t.Errorf("Write(c, %q): got %q, want %q", input, got, input)
	}
}

// Verify that writing an escape sequence delegates correctly.
func TestEsc(t *testing.T) {
	const input = "OK"

	var buf bytes.Buffer
	c := NewCoder(&buf)

	if _, err := Esc(c, 'a', "A-"); err != nil {
		t.Fatalf("Esc: unexpected error: %v", err)
	}
	io.WriteString(c, input)
	if got, want := buf.String(), "\033aA-OK"; got != want {
		t.Errorf(`c.Esc(a, 'a', "A-").Write(%q): got %q, want %q`, input, got, want)
	}
}

// Verify that wrapped writes are delegated correctly.
func TestSet(t *testing.T) {
	tests := []struct {
		start, end     byte
		prefix, suffix string
		input, want    string
	}{
		// No trailer, no prefix, no suffix.
		{'a', 0, "", "", "OK", "\033aOK"},

		// Trailer, no prefix, no suffix.
		{'a', 'b', "", "", "OK", "\033aOK\033b"},

		// Suffix, no prefix, no trailer.
		{'a', 0, "", "cool", "OK", "\033aOKcool"},

		// Suffix and trailer, no prefix.
		{'a', 'b', "", "bye", "OK", "\033aOK\033bbye"},

		// Prefix, no trailer, no suffix.
		{'a', 0, "A-", "", "OK", "\033aA-OK"},

		// Prefix and suffix, no trailer.
		{'a', 0, "A-", "-mate", "OK", "\033aA-OK-mate"},

		// Prefix, suffix, and trailer.
		{'a', 'b', "A-", "-mate", "OK", "\033aA-OK\033b-mate"},
	}
	for _, test := range tests {
		for _, term := range []bool{false, true} {
			// Unconditional sets should always mark up their output.
			t.Run(fmt.Sprintf("Set-%v", term), func(t *testing.T) {
				var buf bytes.Buffer
				c := NewCoder(&buf)
				c.isTerm = term

				w := c.Set(test.start, test.prefix, test.end, test.suffix)
				io.WriteString(w, test.input)
				if got := buf.String(); got != test.want {
					t.Errorf("Set(%c, %q, %c, %q).Write(%q) term=%v: got %q, want %q",
						test.start, test.prefix, test.end, test.suffix, test.input, term, got, test.want)
				}
			})

			// Conditional sets should only mark up their output if term is true.
			t.Run(fmt.Sprintf("SetIf-%v", term), func(t *testing.T) {
				var buf bytes.Buffer
				c := NewCoder(&buf)
				c.isTerm = term
				want := test.input
				if term {
					want = test.want
				}

				w := c.SetIf(test.start, test.prefix, test.end, test.suffix)
				io.WriteString(w, test.input)
				if got := buf.String(); got != want {
					t.Errorf("SetIf(%c, %q, %c, %q).Write(%q) term=%v: got %q, want %q",
						test.start, test.prefix, test.end, test.suffix, test.input, term, got, want)
				}
			})
		}
	}
}
