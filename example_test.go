/*
 * Package xz examples
 *
 * Authors: Michael Cross <https://xi2.org/x/xz>
 *
 * This file has been put into the public domain.
 * You can do whatever you want with this file.
 */

package xz_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"xi2.org/x/xz"
)

func ExampleReader_Read() {
	data, err := ioutil.ReadFile(
		filepath.Join("testdata", "xz-utils", "good-1-check-sha256.xz"))
	if err != nil {
		log.Fatal(err)
	}
	r, err := xz.NewReader(bytes.NewReader(data), 0)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatal(err)
	}
	// Output:
	// Hello
	// World!
}

func ExampleReader_Multistream() {
	data, err := ioutil.ReadFile(
		filepath.Join("testdata", "xz-utils", "good-1-check-sha256.xz"))
	if err != nil {
		log.Fatal(err)
	}
	br1, br2 := bytes.NewReader(data), bytes.NewReader(data)
	r, err := xz.NewReader(io.MultiReader(br1, br2), 0)
	if err != nil {
		log.Fatal(err)
	}
	r.Multistream(false)
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "Read first stream\n")
	err = r.Reset()
	if err != nil {
		log.Fatal(err)
	}
	r.Multistream(false)
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "Read second stream\n")
	err = r.Reset()
	if err == io.EOF {
		fmt.Fprintf(os.Stdout, "No more streams\n")
	}
	// Output:
	// Hello
	// World!
	// Read first stream
	// Hello
	// World!
	// Read second stream
	// No more streams
}
