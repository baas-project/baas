// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package images defines the models representing different image types
package images

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"gorm.io/gorm"
)

// FilePathFmt is the format string to create the path the image should be written to
const FilePathFmt = "/%s/%v.img"

// DiskType describes the type of disk image, this can also describe the filesystem contained within
type DiskType int

const (
	// DiskTypeRaw is the simplest DiskType of which nothing extra is known
	DiskTypeRaw DiskType = iota
	// DiskTypeQCow2 defines an image of the QCow type used by qemu
	DiskTypeQCow2
)

// String returns a string associated with a DiskType
func (s DiskType) String() string {
	return toString[s]
}

var toString = map[DiskType]string{
	DiskTypeRaw:   "raw",
	DiskTypeQCow2: "qcow2",
}

var toID = map[string]DiskType{
	"raw":   DiskTypeRaw,
	"qcow2": DiskTypeQCow2,
}

// DiskCompressionStrategy is how the disk is compressed
type DiskCompressionStrategy string

const (
	// DiskCompressionStrategyNone doesn't compress
	DiskCompressionStrategyNone DiskCompressionStrategy = "none"
	// DiskCompressionStrategyZSTD compresses disk images with zstd.
	DiskCompressionStrategyZSTD = "zstd"
	// DiskCompressionStrategyGZip uses the standard GZip compression algorithm for disks.
	DiskCompressionStrategyGZip = "GZip"
)

// MarshalJSON marshals the enum as a quoted json string
func (s DiskType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *DiskType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*s = toID[j]
	return nil
}

// ImageUUID is a UUID distinguishing each disk image
type ImageUUID string

// Version stores the version of an ImageModel using an UNIX timestamp
type Version struct {
	gorm.Model     `json:"-"`
	Version        uint64    `gorm:"not null;default:0"`
	ImageModelUUID ImageUUID `gorm:"not null;"`
}

/* Disk Layout on control_server
/disks
	/abc  <-- First image UUID
		/1.img
		/2.img
		/3.img
		/4.img
	/cdf  <-- Second image UUID
		/1.img
		/2.img
*/

// ImageModel defines the database structure for storing the metadata about images
type ImageModel struct {
	// You will see quite a few of these around. They suppress the default values that the ORM creates when it gets
	// cast into JSON.

	// Human identifiable name of this image
	Name string `gorm:"not null"`

	// Versions are all possible versions of this image, represented as unix
	// timestamps of their creation. A new version is created whenever a reprovisioning
	// takes place, and this image is replaced.
	Versions []Version `gorm:"foreignKey:ImageModelUUID;not null;references:UUID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`

	// ImageUUID is a universally unique identifier for images
	UUID ImageUUID `gorm:"uniqueIndex;primaryKey;unique"`

	// Foreign key for gorm
	Username string `gorm:"foreignKey:Username;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`

	// Compression algorithm used for this image
	DiskCompressionStrategy DiskCompressionStrategy `gorm:"not null;"`

	// The Image Filetype
	ImageFileType DiskType `gorm:"not null;"`

	// The image type
	Type string `gorm:"not null;"`

	// Checksum for this image as alternative for versioning
	Checksum string
}

const (
	// SizeMegabyte are the bytes equivalent to one megabyte
	SizeMegabyte uint = 1024 * 1024
	// SizeGigabyte are the bytes equivalent to one gigabyte
	SizeGigabyte = 1024 * 1024 * 1024
)

// CreateImageFile creates the actual image on disk with a given size.
func (image ImageModel) CreateImageFile(imageSize uint, diskpath string, baseSize uint) error {
	f, err := os.OpenFile(fmt.Sprintf(diskpath+FilePathFmt, image.UUID, "0"),
		os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return err
	}

	// Create an image of a specified size in GiB
	size := int64(imageSize * baseSize)

	_, err = f.Seek(size-1, 0)
	if err != nil {
		return err
	}

	_, err = f.Write([]byte{0})
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
