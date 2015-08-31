/*
 * XZ decompressor
 *
 * Authors: Lasse Collin <lasse.collin@tukaani.org>
 *          Igor Pavlov <http://7-zip.org/>
 *
 * Translation to Go: Michael Cross <https://xi2.org/x/xz>
 *
 * This file has been put into the public domain.
 * You can do whatever you want with this file.
 */

package xz

import (
	"hash/crc32"
	"hash/crc64"
)

/* from linux/include/linux/xz.h **************************************/

/**
 * xzRet - Return codes
 * @xzOK:                   Everything is OK so far. More input or more
 *                          output space is required to continue.
 * @xzStreamEnd:            Operation finished successfully.
 * @xzUnSupportedCheck:     Integrity check type is not supported. Decoding
 *                          is still possible by simply calling xzDecRun
 *                          again.
 * @xzMemlimitError:        A bigger LZMA2 dictionary would be needed than
 *                          allowed by the dictMax argument given to
 *                          xzDecInit.
 * @xzFormatError:          File format was not recognized (wrong magic
 *                          bytes).
 * @xzOptionsError:         This implementation doesn't support the requested
 *                          compression options. In the decoder this means
 *                          that the header CRC32 matches, but the header
 *                          itself specifies something that we don't support.
 * @xzDataError:            Compressed data is corrupt.
 * @xzBufError:             Cannot make any progress.
 *
 * xzBufError is returned when two consecutive calls to XZ code cannot
 * consume any input and cannot produce any new output.  This happens
 * when there is no new input available, or the output buffer is full
 * while at least one output byte is still pending. Assuming your code
 * is not buggy, you can get this error only when decoding a
 * compressed stream that is truncated or otherwise corrupt.
 */
type xzRet int

const (
	xzOK xzRet = iota
	xzStreamEnd
	xzUnsupportedCheck
	xzMemlimitError
	xzFormatError
	xzOptionsError
	xzDataError
	xzBufError
)

/**
 * xzBuf - Passing input and output buffers to XZ code
 * @in:         Input buffer.
 * @inPos:      Current position in the input buffer. This must not exceed
 *              input buffer size.
 * @out:        Output buffer.
 * @outPos:     Current position in the output buffer. This must not exceed
 *              output buffer size.
 *
 * Only the contents of the output buffer from out[outPos] onward, and
 * the variables inPos and outPos are modified by the XZ code.
 */
type xzBuf struct {
	in     []byte
	inPos  int
	out    []byte
	outPos int
}

var xzCRC32Table = crc32.MakeTable(crc32.IEEE)
var xzCRC64Table = crc64.MakeTable(crc64.ECMA)

/*
 * Update CRC32 value using the polynomial from IEEE-802.3. To start a new
 * calculation, the second argument must be zero. To continue the calculation,
 * the previously returned value is passed as the second argument.
 */
func xzCRC32(buf []byte, crc uint32) uint32 {
	return crc32.Update(crc, xzCRC32Table, buf)
}

/* All XZ filter IDs */
type xzFilterID byte

const (
	idDelta       xzFilterID = 0x03
	idBCJX86      xzFilterID = 0x04
	idBCJPowerPC  xzFilterID = 0x05
	idBCJIA64     xzFilterID = 0x06
	idBCJARM      xzFilterID = 0x07
	idBCJARMThumb xzFilterID = 0x08
	idBCJSPARC    xzFilterID = 0x09
	idLZMA2       xzFilterID = 0x21
)
