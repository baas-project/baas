// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package compression defines the methods used to compress and decompress file streams.
package compression

import (
	"io"
	"strings"

	"github.com/baas-project/baas/pkg/model/images"

	gzip "github.com/klauspost/pgzip"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/gozstd"
)

// browbeat forces the strategy to always conform to a lower case
// version of the name. This is done since the JSON library can
// flip to title case which breaks the program. Do not ask me
// why a library can convert to an invalid enum value.
func browbeat(strategy images.DiskCompressionStrategy) images.DiskCompressionStrategy {
	return images.DiskCompressionStrategy(strings.ToLower(string(strategy)))
}

// Decompress is a decorator to decompress a disk image stream
func Decompress(reader io.Reader, strategy images.DiskCompressionStrategy) (io.Reader, error) {
	log.Infof("DiskUUID compression strategy: %v", strategy)
	switch browbeat(strategy) {
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
	log.Infof("DiskUUID compression strategy: %v", strategy)

	switch browbeat(strategy) {
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
