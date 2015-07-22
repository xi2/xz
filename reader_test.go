/*
 * Package xz tests
 *
 * Author: Michael Cross <https://xi2.org/x/xz>
 *
 * This file has been put into the public domain.
 * You can do whatever you want with this file.
 */

package xz_test

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"xi2.org/x/xz"
)

type testFile struct {
	file   string
	md5sum string
	err    error
}

// Note that the md5sums below were generated with XZ Utils and
// XZ Embedded, not with this package.

var badFiles = []testFile{
	{
		file:   "bad-0-backward_size.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-0cat-alone.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrFormat,
	},
	{
		file:   "bad-0cat-header_magic.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrFormat,
	},
	{
		file:   "bad-0catpad-empty.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-0-empty-truncated.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrBuf,
	},
	{
		file:   "bad-0-footer_magic.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-0-header_magic.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrFormat,
	},
	{
		file:   "bad-0-nonempty_index.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-0pad-empty.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrBuf,
	},
	{
		file:   "bad-1-block_header-1.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-block_header-2.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-block_header-3.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-block_header-4.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-block_header-5.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-block_header-6.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-check-crc32.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-check-crc64.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-check-sha256.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-lzma2-1.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-lzma2-2.xz",
		md5sum: "211dbb3d39f3c244585397f6d3c09be3",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-lzma2-3.xz",
		md5sum: "211dbb3d39f3c244585397f6d3c09be3",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-lzma2-4.xz",
		md5sum: "6492b8d167aee3ca222d07a49d24015a",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-lzma2-5.xz",
		md5sum: "6492b8d167aee3ca222d07a49d24015a",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-lzma2-6.xz",
		md5sum: "09f7e02f1290be211da707a266f153b3",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-lzma2-7.xz",
		md5sum: "c214a5e586cb3f0673cc6138f7de25ab",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-lzma2-8.xz",
		md5sum: "211dbb3d39f3c244585397f6d3c09be3",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-stream_flags-1.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-stream_flags-2.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-stream_flags-3.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-vli-1.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-1-vli-2.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrData,
	},
	{
		file:   "bad-2-compressed_data_padding.xz",
		md5sum: "09f7e02f1290be211da707a266f153b3",
		err:    xz.ErrData,
	},
	{
		file:   "bad-2-index-1.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-2-index-2.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-2-index-3.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-2-index-4.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
	{
		file:   "bad-2-index-5.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    xz.ErrData,
	},
}

var goodFiles = []testFile{
	{
		file:   "good-0cat-empty.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    nil,
	},
	{
		file:   "good-0catpad-empty.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    nil,
	},
	{
		file:   "good-0-empty.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    nil,
	},
	{
		file:   "good-0pad-empty.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    nil,
	},
	{
		file:   "good-1-3delta-lzma2.xz",
		md5sum: "c214a5e586cb3f0673cc6138f7de25ab",
		err:    nil,
	},
	{
		file:   "good-1-block_header-1.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    nil,
	},
	{
		file:   "good-1-block_header-2.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    nil,
	},
	{
		file:   "good-1-block_header-3.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    nil,
	},
	{
		file:   "good-1-check-crc32.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    nil,
	},
	{
		file:   "good-1-check-crc64.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    nil,
	},
	{
		file:   "good-1-check-none.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    nil,
	},
	{
		file:   "good-1-check-sha256.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    nil,
	},
	{
		file:   "good-1-delta-lzma2.tiff.xz",
		md5sum: "c692be6d1987989af5eeafc329085ad2",
		err:    nil,
	},
	{
		file:   "good-1-lzma2-1.xz",
		md5sum: "c214a5e586cb3f0673cc6138f7de25ab",
		err:    nil,
	},
	{
		file:   "good-1-lzma2-2.xz",
		md5sum: "c214a5e586cb3f0673cc6138f7de25ab",
		err:    nil,
	},
	{
		file:   "good-1-lzma2-3.xz",
		md5sum: "c214a5e586cb3f0673cc6138f7de25ab",
		err:    nil,
	},
	{
		file:   "good-1-lzma2-4.xz",
		md5sum: "c214a5e586cb3f0673cc6138f7de25ab",
		err:    nil,
	},
	{
		file:   "good-1-lzma2-5.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    nil,
	},
	{
		file:   "good-1-sparc-lzma2.xz",
		md5sum: "835f2865f1d7c7ad2c7de0d5fd07faef",
		err:    nil,
	},
	{
		file:   "good-1-x86-lzma2.xz",
		md5sum: "ce212d6a1cfe73d8395a2b42f94c2419",
		err:    nil,
	},
	{
		file:   "good-2-lzma2.xz",
		md5sum: "fbf68a8e34b2ded53bba54e68794b4fe",
		err:    nil,
	},
}

var unsupportedFiles = []testFile{
	{
		file:   "unsupported-block_header.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrOptions,
	},
	{
		file:   "unsupported-check.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrUnsupportedCheck,
	},
	{
		file:   "unsupported-filter_flags-1.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrOptions,
	},
	{
		file:   "unsupported-filter_flags-2.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrOptions,
	},
	{
		file:   "unsupported-filter_flags-3.xz",
		md5sum: "d41d8cd98f00b204e9800998ecf8427e",
		err:    xz.ErrOptions,
	},
}

var otherFiles = []testFile{
	{
		file:   "good-1-x86-lzma2-offset-2048.xz",
		md5sum: "ce212d6a1cfe73d8395a2b42f94c2419",
		err:    nil,
	},
	{
		file:   "good-2-lzma2-corrupt.xz",
		md5sum: "d9c5223e7e6e305e6c1c6ed73789df88",
		err:    xz.ErrData,
	},
	{
		file:   "pg2242.txt.xz",
		md5sum: "0b4e9c76f4dff77d72f541e5b2845f5e",
		err:    nil,
	},
	{
		file:   "random-1mb.xz",
		md5sum: "3f04b090e5d26a1cbeea53c21ebcad03",
		err:    nil,
	},
	{
		file:   "zeros-100mb.xz",
		md5sum: "0f86d7c5a6180cf9584c1d21144d85b0",
		err:    nil,
	},
}

func testFileList(t *testing.T, dir string, files []testFile) {
	for _, f := range files {
		func() {
			fr, err := os.Open(filepath.Join("testdata", dir, f.file))
			if err != nil {
				t.Fatal(err)
			}
			defer fr.Close()
			hash := md5.New()
			r, err := xz.NewReader(fr, 0)
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(hash, r)
			if err != f.err {
				t.Fatalf("%s: wanted error: %v, got: %v\n", f.file, f.err, err)
			}
			md5sum := fmt.Sprintf("%x", hash.Sum(nil))
			if f.md5sum != md5sum {
				t.Fatalf(
					"%s: wanted md5: %v, got: %v\n", f.file, f.md5sum, md5sum)
			}
		}()
	}
}

func TestBadFiles(t *testing.T) {
	testFileList(t, "xz-utils", badFiles)
}

func TestGoodFiles(t *testing.T) {
	testFileList(t, "xz-utils", goodFiles)
}

func TestUnsupportedFiles(t *testing.T) {
	testFileList(t, "xz-utils", unsupportedFiles)
}

func TestOtherFiles(t *testing.T) {
	testFileList(t, "other", otherFiles)
}

func TestMemlimit(t *testing.T) {
	data, err := ioutil.ReadFile(
		filepath.Join("testdata", "other", "pg2242.txt.xz"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := xz.NewReader(bytes.NewReader(data), 1<<25)
	if err != nil {
		t.Fatal(err)
	}
	b := new(bytes.Buffer)
	_, err = io.Copy(b, r)
	if err != xz.ErrMemlimit {
		t.Fatalf("wanted error: %v, got: %v\n", xz.ErrMemlimit, err)
	}
}

// test to ensure that decoder errors are not returned prematurely
// the test file returns 6 decoded bytes before corruption occurs
func TestPrematureError(t *testing.T) {
	data, err := ioutil.ReadFile(
		filepath.Join("testdata", "other", "good-2-lzma2-corrupt.xz"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := xz.NewReader(bytes.NewReader(data), 0)
	if err != nil {
		t.Fatal(err)
	}
	b := make([]byte, 2)
	n, err := r.Read(b)
	if n != 2 || err != nil {
		t.Fatalf("Read returned: (%d,%v), expected: (2,%v)\n", n, err, nil)
	}
	n, err = r.Read(b)
	if n != 2 || err != nil {
		t.Fatalf("Read returned: (%d,%v), expected: (2,%v)\n", n, err, nil)
	}
	n, err = r.Read(b)
	if n != 2 || err != xz.ErrData {
		t.Fatalf("Read returned: (%d,%v), expected: (2,%v)\n",
			n, err, xz.ErrData)
	}
}

func TestMultipleBadReads(t *testing.T) {
	data, err := ioutil.ReadFile(
		filepath.Join("testdata", "other", "good-2-lzma2-corrupt.xz"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := xz.NewReader(bytes.NewReader(data), 0)
	if err != nil {
		t.Fatal(err)
	}
	b := make([]byte, 100)
	n, err := r.Read(b)
	if n != 6 || err != xz.ErrData {
		t.Fatalf("Read returned: (%d,%v), expected: (6,%v)\n",
			n, err, xz.ErrData)
	}
	n, err = r.Read(b)
	if n != 0 || err != xz.ErrData {
		t.Fatalf("Read returned: (%d,%v), expected: (0,%v)\n",
			n, err, xz.ErrData)
	}
	n, err = r.Read(b)
	if n != 0 || err != xz.ErrData {
		t.Fatalf("Read returned: (%d,%v), expected: (0,%v)\n",
			n, err, xz.ErrData)
	}
}

// Multistream is tested in example_test.go
