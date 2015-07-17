// Package xz implements XZ decompression natively in Go.
//
// Usage
//
// For ease of use, this package is designed to have a similar
// interface to compress/gzip. See the examples for more details.
//
// Implementation details and limitations
//
// This package is a translation from C to Go of XZ Embedded
// <http://tukaani.org/xz/embedded.html> with small enhancements, and
// as such shares some (but not all) of its limitations. It cannot
// decompress every conceivable XZ file out there. However, it should
// decompress the vast majority of files one is likely to come
// across. In particular, if you are using files created with XZ Utils
// "xz -9" with default settings, or with GNU Tar "tar -J", you should
// be fine.
//
// The technical limitations, shared with XZ Embedded, are as follows:
//
//   - Only LZMA2 and BCJ filters are supported (no Delta filter).
//   - A filter chain may consist of just a single LZMA2 filter or
//     single BCJ + single LZMA2.
//   - BCJ filters must have a zero start offset.
//
// Package xz has the following enhancements over XZ Embedded:
//
//   - All known block check types (including SHA256) are implemented.
//   - Correct handling of multiple streams and stream padding.
//
// See <http://tukaani.org/xz/xz-file-format.txt> for the specifics of
// the XZ file format.
//
// For bug reports relating to this package please contact the author
// through <https://xi2.org/x/xz> and not the author of XZ Embedded.
package xz
