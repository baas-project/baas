package images

type FilesystemType string

const (
	FileSystemTypeFAT32 FilesystemType = "fat32"
	FilesystemEXT4                     = "ext4"
)

type MachineModel struct {
	ImageModel
	machineID  uint
	filesystem FilesystemType
	size       uint // filesize in MiB
}

/*
func (m MachineModel) create(machine model.MachineModel) (*MachineModel, error) {
	uuId, err := uuid.NewUUID()

	if err != nil {
		return nil, err
	}

	image := ImageModel{
		Versions: []Version{},
		UUID:     ImageUUID(uuId.String()),
	}

	machineImage := MachineModel{ImageModel: image,
		machineID:  machine.ID,
		size:       128,
		filesystem: FilesystemEXT4,
	}

	machineImage.CreateImageFile(machineImage.size, "/tmp/test.img", SizeMegabyte)

	return &machineImage, nil
}
*/
