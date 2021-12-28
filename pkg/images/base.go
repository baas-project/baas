package images

import (
	"bytes"
	"encoding/json"
	"gorm.io/gorm"
)

// DiskType describes the type of disk image, this can also describe the filesystem contained within
type DiskType int

const (
	// DiskTypeRaw is the simplest DiskType of which nothing extra is known
	DiskTypeRaw DiskType = iota
	DiskTypeQCow2
)

func (s DiskType) String() string {
	return toString[s]
}

var toString = map[DiskType]string{
	DiskTypeRaw:   "raw",
	DiskTypeQCow2: "qcow2",
}

var toId = map[string]DiskType{
	"raw":   DiskTypeRaw,
	"qcow2": DiskTypeQCow2,
}

type DiskCompressionStrategy string

const (
	// DiskCompressionStrategyNone doesn't compress
	DiskCompressionStrategyNone DiskCompressionStrategy = "none"
	// DiskCompressionStrategyZSTD compresses disk images with zstd.
	DiskCompressionStrategyZSTD = "zstd"
	// DiskCompressionStrategyGZip uses the standard GZip compression algorithm for disks.
	DiskCompressionStrategyGZip = "GZip"
)

// DiskImage describes a single disk image on the machine
type DiskImage struct {
	gorm.Model `json:"-"`

	DiskType                DiskType
	DiskCompressionStrategy DiskCompressionStrategy
}

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
	*s = toId[j]
	return nil
}

// ImageUUID is a UUID distinguishing each disk image
type ImageUUID string

// Version stores the version of an ImageModel using an UNIX timestamp
type Version struct {
	gorm.Model `json:"-"`

	Version      uint64 `gorm:"autoIncrement;primaryKey;not null;unique"`
	ImageModelID uint
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
	gorm.Model `json:"-"`

	// Human identifiable name of this image
	Name string

	// Versions are all possible versions of this image, represented as unix
	// timestamps of their creation. A new version is created whenever a reprovisioning
	// takes place, and this image is replaced.
	Versions []Version

	// ImageUUID is a universally unique identifier for images
	UUID ImageUUID `gorm:"uniqueIndex;primaryKey;unique"`

	// Foreign key for gorm
	UserModelID uint32
}
