package main

import (
	"io"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/gozstd"

	"baas/pkg/model"
)

// Decompress is a decorator to decompress a disk image stream
func Decompress(reader io.Reader, image model.DiskImage) (io.Reader, error) {
	log.Debugf("Disk compression strategy: %v", image.DiskCompressionStrategy)
	switch image.DiskCompressionStrategy {
	case model.DiskCompressionStrategyNone:
		return reader, nil
	case model.DiskCompressionStrategyZSTD:
		newreader := gozstd.NewReader(reader)
		return newreader, nil
	default:
		return nil, errors.New("unknown decompression strategy")
	}
}

// Compress is a decorator to compress a disk image stream
func Compress(reader io.Reader, image model.DiskImage) (io.Reader, error) {
	log.Debugf("Disk compression strategy: %v", image.DiskCompressionStrategy)

	switch image.DiskCompressionStrategy {
	case model.DiskCompressionStrategyNone:
		return reader, nil
	case model.DiskCompressionStrategyZSTD:
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
