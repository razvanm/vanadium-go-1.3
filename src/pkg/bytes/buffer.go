// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

// Simple byte buffer for marshaling data.

import (
	"io"
	"os"
	"utf8"
)

// A Buffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer struct {
	buf       []byte            // contents are the bytes buf[off : len(buf)]
	off       int               // read at &buf[off], write at &buf[len(buf)]
	runeBytes [utf8.UTFMax]byte // avoid allocation of slice on each WriteByte or Rune
	bootstrap [64]byte          // memory to hold first slice; helps small buffers (Printf) avoid allocation.
	lastRead  readOp            // last read operation, so that Unread* can work correctly.
}

// The readOp constants describe the last action performed on
// the buffer, so that UnreadRune and UnreadByte can
// check for invalid usage.
type readOp int

const (
	opInvalid  readOp = iota // Non-read operation.
	opReadRune               // Read rune.
	opRead                   // Any other read operation.
)

// Bytes returns a slice of the contents of the unread portion of the buffer;
// len(b.Bytes()) == b.Len().  If the caller changes the contents of the
// returned slice, the contents of the buffer will change provided there
// are no intervening method calls on the Buffer.
func (b *Buffer) Bytes() []byte { return b.buf[b.off:] }

// String returns the contents of the unread portion of the buffer
// as a string.  If the Buffer is a nil pointer, it returns "<nil>".
func (b *Buffer) String() string {
	if b == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	return string(b.buf[b.off:])
}

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (b *Buffer) Len() int { return len(b.buf) - b.off }

// Truncate discards all but the first n unread bytes from the buffer.
// It is an error to call b.Truncate(n) with n > b.Len().
func (b *Buffer) Truncate(n int) {
	b.lastRead = opInvalid
	if n == 0 {
		// Reuse buffer space.
		b.off = 0
	}
	b.buf = b.buf[0 : b.off+n]
}

// Reset resets the buffer so it has no content.
// b.Reset() is the same as b.Truncate(0).
func (b *Buffer) Reset() { b.Truncate(0) }

// Grow buffer to guarantee space for n more bytes.
// Return index where bytes should be written.
func (b *Buffer) grow(n int) int {
	m := b.Len()
	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Truncate(0)
	}
	if len(b.buf)+n > cap(b.buf) {
		var buf []byte
		if b.buf == nil && n <= len(b.bootstrap) {
			buf = b.bootstrap[0:]
		} else {
			// not enough space anywhere
			buf = make([]byte, 2*cap(b.buf)+n)
			copy(buf, b.buf[b.off:])
		}
		b.buf = buf
		b.off = 0
	}
	b.buf = b.buf[0 : b.off+m+n]
	return b.off + m
}

// Write appends the contents of p to the buffer.  The return
// value n is the length of p; err is always nil.
func (b *Buffer) Write(p []byte) (n int, err os.Error) {
	b.lastRead = opInvalid
	m := b.grow(len(p))
	copy(b.buf[m:], p)
	return len(p), nil
}

// WriteString appends the contents of s to the buffer.  The return
// value n is the length of s; err is always nil.
func (b *Buffer) WriteString(s string) (n int, err os.Error) {
	b.lastRead = opInvalid
	m := b.grow(len(s))
	return copy(b.buf[m:], s), nil
}

// MinRead is the minimum slice size passed to a Read call by
// Buffer.ReadFrom.  As long as the Buffer has at least MinRead bytes beyond
// what is required to hold the contents of r, ReadFrom will not grow the
// underlying buffer.
const MinRead = 512

// ReadFrom reads data from r until EOF and appends it to the buffer.
// The return value n is the number of bytes read.
// Any error except os.EOF encountered during the read
// is also returned.
func (b *Buffer) ReadFrom(r io.Reader) (n int64, err os.Error) {
	b.lastRead = opInvalid
	// If buffer is empty, reset to recover space.
	if b.off >= len(b.buf) {
		b.Truncate(0)
	}
	for {
		if cap(b.buf)-len(b.buf) < MinRead {
			var newBuf []byte
			// can we get space without allocation?
			if b.off+cap(b.buf)-len(b.buf) >= MinRead {
				// reuse beginning of buffer
				newBuf = b.buf[0 : len(b.buf)-b.off]
			} else {
				// not enough space at end; put space on end
				newBuf = make([]byte, len(b.buf)-b.off, 2*(cap(b.buf)-b.off)+MinRead)
			}
			copy(newBuf, b.buf[b.off:])
			b.buf = newBuf
			b.off = 0
		}
		m, e := r.Read(b.buf[len(b.buf):cap(b.buf)])
		b.buf = b.buf[0 : len(b.buf)+m]
		n += int64(m)
		if e == os.EOF {
			break
		}
		if e != nil {
			return n, e
		}
	}
	return n, nil // err is EOF, so return nil explicitly
}

// WriteTo writes data to w until the buffer is drained or an error
// occurs. The return value n is the number of bytes written.
// Any error encountered during the write is also returned.
func (b *Buffer) WriteTo(w io.Writer) (n int64, err os.Error) {
	b.lastRead = opInvalid
	for b.off < len(b.buf) {
		m, e := w.Write(b.buf[b.off:])
		n += int64(m)
		b.off += m
		if e != nil {
			return n, e
		}
	}
	// Buffer is now empty; reset.
	b.Truncate(0)
	return
}

// WriteByte appends the byte c to the buffer.
// The returned error is always nil, but is included
// to match bufio.Writer's WriteByte.
func (b *Buffer) WriteByte(c byte) os.Error {
	b.lastRead = opInvalid
	m := b.grow(1)
	b.buf[m] = c
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode
// code point r to the buffer, returning its length and
// an error, which is always nil but is included
// to match bufio.Writer's WriteRune.
func (b *Buffer) WriteRune(r int) (n int, err os.Error) {
	if r < utf8.RuneSelf {
		b.WriteByte(byte(r))
		return 1, nil
	}
	n = utf8.EncodeRune(b.runeBytes[0:], r)
	b.Write(b.runeBytes[0:n])
	return n, nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained.  The return value n is the number of bytes read.  If the
// buffer has no data to return, err is os.EOF even if len(p) is zero;
// otherwise it is nil.
func (b *Buffer) Read(p []byte) (n int, err os.Error) {
	b.lastRead = opInvalid
	if b.off >= len(b.buf) {
		// Buffer is empty, reset to recover space.
		b.Truncate(0)
		return 0, os.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	if n > 0 {
		b.lastRead = opRead
	}
	return
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by Read.
// If there are fewer than n bytes in the buffer, Next returns the entire buffer.
// The slice is only valid until the next call to a read or write method.
func (b *Buffer) Next(n int) []byte {
	b.lastRead = opInvalid
	m := b.Len()
	if n > m {
		n = m
	}
	data := b.buf[b.off : b.off+n]
	b.off += n
	if n > 0 {
		b.lastRead = opRead
	}
	return data
}

// ReadByte reads and returns the next byte from the buffer.
// If no byte is available, it returns error os.EOF.
func (b *Buffer) ReadByte() (c byte, err os.Error) {
	b.lastRead = opInvalid
	if b.off >= len(b.buf) {
		// Buffer is empty, reset to recover space.
		b.Truncate(0)
		return 0, os.EOF
	}
	c = b.buf[b.off]
	b.off++
	b.lastRead = opRead
	return c, nil
}

// ReadRune reads and returns the next UTF-8-encoded
// Unicode code point from the buffer.
// If no bytes are available, the error returned is os.EOF.
// If the bytes are an erroneous UTF-8 encoding, it
// consumes one byte and returns U+FFFD, 1.
func (b *Buffer) ReadRune() (r int, size int, err os.Error) {
	b.lastRead = opInvalid
	if b.off >= len(b.buf) {
		// Buffer is empty, reset to recover space.
		b.Truncate(0)
		return 0, 0, os.EOF
	}
	b.lastRead = opReadRune
	c := b.buf[b.off]
	if c < utf8.RuneSelf {
		b.off++
		return int(c), 1, nil
	}
	r, n := utf8.DecodeRune(b.buf[b.off:])
	b.off += n
	return r, n, nil
}

// UnreadRune unreads the last rune returned by ReadRune.
// If the most recent read or write operation on the buffer was
// not a ReadRune, UnreadRune returns an error.  (In this regard
// it is stricter than UnreadByte, which will unread the last byte
// from any read operation.)
func (b *Buffer) UnreadRune() os.Error {
	if b.lastRead != opReadRune {
		return os.ErrorString("bytes.Buffer: UnreadRune: previous operation was not ReadRune")
	}
	b.lastRead = opInvalid
	if b.off > 0 {
		_, n := utf8.DecodeLastRune(b.buf[0:b.off])
		b.off -= n
	}
	return nil
}

// UnreadByte unreads the last byte returned by the most recent
// read operation.  If write has happened since the last read, UnreadByte
// returns an error.
func (b *Buffer) UnreadByte() os.Error {
	if b.lastRead != opReadRune && b.lastRead != opRead {
		return os.ErrorString("bytes.Buffer: UnreadByte: previous operation was not a read")
	}
	b.lastRead = opInvalid
	if b.off > 0 {
		b.off--
	}
	return nil
}

// ReadBytes reads until the first occurrence of delim in the input,
// returning a slice containing the data up to and including the delimiter.
// If ReadBytes encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often os.EOF).
// ReadBytes returns err != nil if and only if the returned data does not end in
// delim.
func (b *Buffer) ReadBytes(delim byte) (line []byte, err os.Error) {
	i := IndexByte(b.buf[b.off:], delim)
	size := i + 1 - b.off
	if i < 0 {
		size = len(b.buf) - b.off
		err = os.EOF
	}
	line = make([]byte, size)
	copy(line, b.buf[b.off:])
	return
}

// ReadString reads until the first occurrence of delim in the input,
// returning a string containing the data up to and including the delimiter.
// If ReadString encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often os.EOF).
// ReadString returns err != nil if and only if the returned data does not end
// in delim.
func (b *Buffer) ReadString(delim byte) (line string, err os.Error) {
	bytes, err := b.ReadBytes(delim)
	return string(bytes), err
}

// NewBuffer creates and initializes a new Buffer using buf as its initial
// contents.  It is intended to prepare a Buffer to read existing data.  It
// can also be used to size the internal buffer for writing.  To do that,
// buf should have the desired capacity but a length of zero.
func NewBuffer(buf []byte) *Buffer { return &Buffer{buf: buf} }

// NewBufferString creates and initializes a new Buffer using string s as its
// initial contents.  It is intended to prepare a buffer to read an existing
// string.
func NewBufferString(s string) *Buffer {
	return &Buffer{buf: []byte(s)}
}
