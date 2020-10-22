package main

import (
	"errors"
	"io"
	"os"
)

// Size with which to copy
const blocksize int64 = 512

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
func CopyStream(src io.ReaderAt, dest io.WriterAt) error {
	buff := make([]byte, blocksize)

	for i := int64(0); ; {
		// Read a block to the buffer
		n, errr := src.ReadAt(buff, i)
		if errr != nil && errr != io.EOF {
			return errr
		}

		// Write the block to the dest file
		if dn, errw := dest.WriteAt(buff[:n], i); errw != nil || dn != n {
			if errw == nil {
				return errors.New("partial copy")
			}
			return errw
		}

		if errr == io.EOF {
			return nil
		}

		i += int64(n)
	}
}
