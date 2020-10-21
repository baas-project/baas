package main

import (
	"errors"
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

	stat, err := src.Stat()
	if err != nil {
		return err
	}

	size := stat.Size()

	buff := make([]byte, blocksize)

	for i := int64(0); i < size; {
		n, err := src.ReadAt(buff, blocksize)
		if err != nil {
			return err
		}
		if dn, err := dest.WriteAt(buff, i); err != nil || dn != n {
			if err != nil {
				return err
			} else {
				return errors.New("partial copy")
			}
		}

		i += int64(n)
	}

	return nil
}
