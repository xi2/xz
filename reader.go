/*
 * Package xz Go Reader API
 *
 * Author: Michael Cross <https://xi2.org/x/xz>
 *
 * This file has been put into the public domain.
 * You can do whatever you want with this file.
 */

package xz // import "xi2.org/x/xz"

import (
	"errors"
	"io"
)

var (
	ErrUnsupportedCheck = errors.New("xz: integrity check type not supported")
	ErrMemlimit         = errors.New("xz: LZMA2 dictionary size exceeds max")
	ErrFormat           = errors.New("xz: file format not recognized")
	ErrOptions          = errors.New("xz: compression options not supported")
	ErrData             = errors.New("xz: data is corrupt")
	ErrBuf              = errors.New("xz: data is truncated or corrupt")
	ErrNilReader        = errors.New("xz: source reader is nil")
)

// DefaultDictMax is the default maximum dictionary size in bytes used
// by the decoder. This value is sufficient to decompress files
// created with XZ Utils "xz -9".
const DefaultDictMax = 1 << 26 // 64 MiB

// bufSize is the input/output buffer size used by the decoder.
const bufSize = 1 << 13 // 8 KiB

// NewReader creates a new Reader reading from r. The decompressor
// will use an LZMA2 dictionary size up to dictMax bytes in
// size. Passing a value of zero sets dictMax to DefaultDictMax.  If
// an individual XZ stream requires a dictionary size greater than
// dictMax in order to decompress, Read will return ErrMemlimit.
//
// Due to internal buffering, the Reader may read more data than
// necessary from r.
func NewReader(r io.Reader, dictMax uint32) (*Reader, error) {
	if r == nil {
		return nil, ErrNilReader
	}
	if dictMax == 0 {
		dictMax = DefaultDictMax
	}
	return &Reader{
		r:           r,
		multistream: true,
		buf: &xzBuf{
			out: make([]byte, bufSize),
		},
		padding: -1,
		dec:     xzDecInit(dictMax),
	}, nil
}

// A Reader is an io.Reader that can be used to retrieve uncompressed
// data from an XZ file.
//
// In general, an XZ file can be a concatenation of other XZ
// files. Reads from the Reader return the concatenation of the
// uncompressed data of each.
type Reader struct {
	r           io.Reader     // the wrapped io.Reader
	multistream bool          // true if reader is in multistream mode
	rEOF        bool          // true after io.EOF received on r
	dEOF        bool          // true after decoder has completed
	padding     int           // bytes of stream padding read (or -1)
	in          [bufSize]byte // backing array for buf.in
	outPos      int           // pos within buf.out of unwritten data
	buf         *xzBuf        // decoder input/output buffers
	dec         *xzDec        // decoder state
	err         error         // the result of the last decoder call
}

// decode is a wrapper around xzDecRun that additionally handles
// stream padding. It treats the padding as a kind of stream that
// decodes to nothing.
//
// When decoding padding, r.padding >= 0
// When decoding a real stream, r.padding == -1
func (r *Reader) decode() (ret xzRet) {
	if r.padding >= 0 {
		// read all padding in input buffer
		for r.buf.inPos < len(r.buf.in) &&
			r.buf.in[r.buf.inPos] == 0 {
			r.buf.inPos++
			r.padding++
		}
		switch {
		case r.buf.inPos == len(r.buf.in) && r.rEOF:
			// case: out of padding. no more input data available
			if r.padding%4 != 0 {
				ret = xzBufError
			} else {
				ret = xzStreamEnd
			}
		case r.buf.inPos == len(r.buf.in):
			// case: read more padding next loop iteration
			ret = xzOK
		default:
			// case: out of padding. more input data available
			if r.padding%4 != 0 {
				ret = xzDataError
			} else {
				xzDecReset(r.dec)
				ret = xzStreamEnd
			}
		}
	} else {
		r.buf.outPos = 0
		r.outPos = 0
		ret = xzDecRun(r.dec, r.buf)
	}
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	// restore err
	err = r.err
	for {
		// copy r.buf.out -> p
		for r.outPos < r.buf.outPos && n < len(p) {
			p[n] = r.buf.out[r.outPos]
			n++
			r.outPos++
		}
		// if p full but output remaining, return with err == nil
		if r.outPos < r.buf.outPos && n == len(p) {
			err = nil
			break
		}
		// all output written. if last call to decoder ended with an
		// error, return that error
		if err != nil {
			break
		}
		// if decoder has finished, return with err == io.EOF
		if r.dEOF {
			err = io.EOF
			break
		}
		// if needed, read more data from r.r
		if r.buf.inPos == len(r.buf.in) && !r.rEOF {
			rn, e := r.r.Read(r.in[:])
			if e != nil && e != io.EOF {
				// read error
				err = e
				break
			}
			if e == io.EOF {
				r.rEOF = true
			}
			// set new input buffer in r.buf
			r.buf.in = r.in[:rn]
			r.buf.inPos = 0
		}
		// decode more data
		ret := r.decode()
		switch ret {
		case xzOK:
			// no action needed
		case xzStreamEnd:
			if r.padding >= 0 {
				r.padding = -1
				if !r.multistream || r.rEOF == true {
					r.dEOF = true
				}
			} else {
				r.padding = 0
			}
		case xzUnsupportedCheck:
			err = ErrUnsupportedCheck
		case xzMemlimitError:
			err = ErrMemlimit
		case xzFormatError:
			err = ErrFormat
		case xzOptionsError:
			err = ErrOptions
		case xzDataError:
			err = ErrData
		case xzBufError:
			err = ErrBuf
		}
		// save err
		r.err = err
	}
	return
}

// Multistream controls whether the reader is operating in multistream
// mode.
//
// If enabled (the default), the Reader expects the input to be a
// sequence of XZ streams, possibly interspersed with stream padding,
// which it reads one after another. The effect is that the
// concatenation of a sequence of XZ streams or XZ files is
// treated as equivalent to the compressed result of the concatenation
// of the sequence. This is standard behaviour for XZ readers.
//
// Calling Multistream(false) disables this behaviour; disabling the
// behaviour can be useful when reading file formats that distinguish
// individual XZ streams. In this mode, when the Reader reaches the
// end of the stream, Read returns io.EOF. To start the next stream,
// call r.Reset() followed by r.Multistream(false). If there is no
// next stream, r.Reset() will return io.EOF.
func (r *Reader) Multistream(ok bool) {
	r.multistream = ok
}

// Reset prepares the reader to read follow on streams when it is not
// in multistream mode and it has finished reading a stream. It also
// resets multistream mode to true (the default). If there are no
// follow on streams, Reset returns io.EOF.
func (r *Reader) Reset() error {
	if !r.dEOF {
		return nil
	}
	if r.rEOF {
		return io.EOF
	}
	r.dEOF = false
	r.multistream = true
	return nil
}
