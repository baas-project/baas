// Package fs provides utility functions for writing and reading from the filesystem in baas.
package fs

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

// Size with which to copy
const blocksize int64 = 1500

// CopyFile is a function which copies a file, it is similar to dd in usage
func CopyFile(from, to string) error {
	src, err := os.OpenFile(from, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}

	dest, err := os.Create(to)
	if err != nil {
		return err
	}

	return CopyStream(src, dest)
}

// CopyStream is a function which copies a stream, it is similar to dd in usage
func CopyStream(src io.Reader, dest io.Writer) error {
	buff := make([]byte, blocksize)

	for {
		// Read a block to the buffer
		n, errr := src.Read(buff)
		if errr != nil && errr != io.EOF && errr != io.ErrUnexpectedEOF {
			return errors.Wrap(errr, "error reading")
		}

		// Write the block to the dest file
		if dn, errw := dest.Write(buff[:n]); errw != nil || dn != n {
			if errw == nil {
				return errors.New("partial copy")
			}
			return errors.Wrap(errw, "error writing")
		}

		if errr == io.EOF {
			return nil
		}
	}
}
