package compression

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/gozstd"

	"github.com/baas-project/baas/pkg/model"
)

func TestCompressZSTD(t *testing.T) {
	b := []byte("Hello, world")
	compressed := gozstd.Compress(nil, b)
	r := bytes.NewReader(compressed)

	decompressed, err := Decompress(r, model.DiskImage{
		DiskCompressionStrategy: model.DiskCompressionStrategyZSTD,
	})
	assert.NoError(t, err)

	res, err := ioutil.ReadAll(decompressed)
	assert.NoError(t, err)

	assert.Equal(t, b, res)
}

func TestDecompressZSTD(t *testing.T) {
	b := []byte("Hello, world")
	r := bytes.NewReader(b)

	c, err := Compress(r, model.DiskImage{
		DiskCompressionStrategy: model.DiskCompressionStrategyZSTD,
	})
	assert.NoError(t, err)

	cb, err := ioutil.ReadAll(c)
	assert.NoError(t, err)

	db, err := gozstd.Decompress(nil, cb)
	assert.NoError(t, err)

	assert.Equal(t, b, db)
}

func TestCompressDecompressZSTD(t *testing.T) {
	b := []byte("Hello, world")
	r := bytes.NewReader(b)

	c, err := Compress(r, model.DiskImage{
		DiskCompressionStrategy: model.DiskCompressionStrategyZSTD,
	})
	assert.NoError(t, err)

	d, err := Decompress(c, model.DiskImage{
		DiskCompressionStrategy: model.DiskCompressionStrategyZSTD,
	})
	assert.NoError(t, err)

	res, err := ioutil.ReadAll(d)
	assert.NoError(t, err)

	assert.Equal(t, b, res)
}
