// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package petrify

import (
	"fmt"
	"io"
)

var (
	newline    = []byte{'\n'}
	dataIndent = []byte{'\t', '\t'}
	space      = []byte{' '}
)

// ByteWriter is used to write bytes of p.
type ByteWriter struct {
	io.Writer
	c int
}

func (w *ByteWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	for n = range p {
		if w.c%12 == 0 {
			_, _ = w.Writer.Write(newline)
			_, _ = w.Writer.Write(dataIndent)
			w.c = 0
		} else {
			_, _ = w.Writer.Write(space)
		}

		_, _ = fmt.Fprintf(w.Writer, "0x%02x,", p[n])
		w.c++
	}

	n++

	return
}
