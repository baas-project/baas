// Package compression defines the methods used to compress and decompress file streams.
package compression

import (
	"github.com/baas-project/baas/pkg/images"
	"io"

	gzip "github.com/klauspost/pgzip"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/gozstd"
)

// Decompress is a decorator to decompress a disk image stream
func Decompress(reader io.Reader, strategy images.DiskCompressionStrategy) (io.Reader, error) {
	log.Debugf("DiskUUID compression strategy: %v", strategy)
	switch strategy {
	case images.DiskCompressionStrategyNone:
		return reader, nil
	case images.DiskCompressionStrategyZSTD:
		return gozstd.NewReader(reader), nil
	default:
		return nil, errors.New("unknown decompression strategy")
	}
}

// Compress is a decorator to compress a disk image stream
func Compress(reader io.Reader, strategy images.DiskCompressionStrategy) (io.Reader, error) {
	log.Debugf("DiskUUID compression strategy: %v", strategy)

	switch strategy {
	case images.DiskCompressionStrategyNone:
		return reader, nil
	case images.DiskCompressionStrategyGZip:
		pr, pw := io.Pipe()
		log.Info("Compress the disk using gunzip")
		go func() {
			w := gzip.NewWriter(pw)
			_, err := io.Copy(w, reader)
			if err != nil {
				log.Warn("Cannot compress data.")
				return
			}
			err = w.Close()
			if err != nil {
				log.Warn("Cannot close gunzip header")
				return
			}
			err = pw.Close()
			if err != nil {
				log.Warn("Close the writing pointer")
				return
			}
		}()

		return pr, nil
	case images.DiskCompressionStrategyZSTD:
		pr, pw := io.Pipe()
		go func() {
			err := gozstd.StreamCompress(pw, reader)
			if err != nil {
				log.Errorf("zstd compression failed")
			}

			err = pw.Close()
			if err != nil {
				log.Errorf("closing pipe failed")
			}
		}()

		return pr, nil
	default:
		return nil, errors.New("unknown decompression strategy")
	}
}
