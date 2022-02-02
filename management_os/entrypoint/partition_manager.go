package main

import (
	"encoding/json"
	"fmt"
	"github.com/baas-project/baas/pkg/images"
	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/part"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

const path = "partitions_cache.json"

var partitionList []Partition

// Partition defines a structure which keeps track of what images are currently on the disk
type Partition struct {
	Partition       part.Partition
	Number          uint32
	AssociatedImage images.ImageUUID
	LastUsedTime    int64
	DeviceFile      string
}

func getPartitions(machine *MachineImage) {
	v, err := machine.Exists(path)
	if err != nil {
		log.Warn("Cannot find existence of cache path: %v", err)
	}

	if !v {
		f, err := machine.Create(path)
		if err != nil {
			log.Warnf("Cannot create partition cache file ")
		}
		partitionList = generatePartitionList()
		writePartitionJson(f)

		err = f.Close()
		if err != nil {
			log.Warnf("Cannot close the partition cache file")
		}
	} else {
		log.Warn("Load cache from file")
		f, err := machine.Open(path)
		if err != nil {
			log.Warnf("Cannot create the partition cache file: %v", err)
		}
		err = json.NewDecoder(f).Decode(&partitionList)
		if err != nil {
			log.Warnf("Cannot read the partition cache: %v", err)
		}
		err = f.Close()
		if err != nil {
			log.Warn("Cannot close the partition cache file")
		}
	}
}
func generatePartitionList() []Partition {
	disk, err := diskfs.OpenWithMode("/dev/sda", diskfs.ReadOnly)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	table, err := disk.GetPartitionTable()

	if err != nil {
		fmt.Printf("%v\n", err)
	}

	var partitionList []Partition
	for i, partition := range table.GetPartitions() {
		if partition.GetSize() != 512 && partition.GetStart() != 0 {
			partitionList = append(partitionList, Partition{
				Partition:       partition,
				Number:          uint32(i + 1),
				AssociatedImage: "",
				LastUsedTime:    time.Now().Unix(),
				DeviceFile:      fmt.Sprintf("/dev/sda%d", uint32(i+1)),
			})
		}
	}

	return partitionList
}

func getPartition(image images.ImageUUID) *Partition {
	chosenPartition := &partitionList[0]

	for i, _ := range partitionList {
		if partitionList[i].AssociatedImage == image {
			log.Info("Found a cached version on disk")
			partitionList[i].LastUsedTime = time.Now().Unix()
			chosenPartition = &partitionList[i]
			break
		}

		if chosenPartition.LastUsedTime > partitionList[i].LastUsedTime {
			chosenPartition = &partitionList[i]
		}
	}

	// Set the partition metadata
	(*chosenPartition).LastUsedTime = time.Now().Unix()
	(*chosenPartition).AssociatedImage = image
	return chosenPartition
}

func writePartitionJson(file *os.File) {
	err := json.NewEncoder(file).Encode(partitionList)

	if err != nil {
		fmt.Printf("Cannot encode JSON to file: %v", err)
	}
}
func printPartitions() {
	for _, partition := range partitionList {
		printPartition(partition)
	}
}

func printPartition(partition Partition) {
	log.Debugf("%s %s %d\n",
		partition.DeviceFile, partition.AssociatedImage,
		partition.LastUsedTime)
}
