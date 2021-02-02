package compression

import (
	"io"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/gozstd"

	"github.com/baas-project/baas/pkg/model"
)

// Decompress is a decorator to decompress a disk image stream
func Decompress(reader io.Reader, strategy model.DiskCompressionStrategy) (io.Reader, error) {
	log.Debugf("DiskUUID compression strategy: %v", strategy)
	switch strategy {
	case model.DiskCompressionStrategyNone:
		return reader, nil
	case model.DiskCompressionStrategyZSTD:
		return gozstd.NewReader(reader), nil
	default:
		return nil, errors.New("unknown decompression strategy")
	}
}

// Compress is a decorator to compress a disk image stream
func Compress(reader io.Reader, strategy model.DiskCompressionStrategy) (io.Reader, error) {
	log.Debugf("DiskUUID compression strategy: %v", strategy)

	switch strategy {
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
