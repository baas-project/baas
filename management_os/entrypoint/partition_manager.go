// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/baas-project/baas/pkg/model/images"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/part"
	log "github.com/sirupsen/logrus"
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

// getPartitions populates the partitionList by either generating the cache or loading from disk
func getPartitions(machine *MachineImage) {
	v, err := machine.Exists(path)
	if err != nil {
		log.Warnf("Cannot find existence of cache path: %v", err)
	}

	if !v {
		f, err := machine.Create(path)
		if err != nil {
			log.Warnf("Cannot create partition cache file: %v", err)
		}
		partitionList = generatePartitionList()
		writePartitionJSON(f)

		err = f.Close()
		if err != nil {
			log.Warnf("Cannot close the partition cache file: %v", err)
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

// generatePartitionList generates a new instance of the cache
func generatePartitionList() []Partition {
	disk, err := diskfs.OpenWithMode("/dev/sda", diskfs.ReadOnly)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	table, err := disk.GetPartitionTable()

	if err != nil {
		fmt.Printf("%v\n", err)
	}

	var partitions []Partition
	for i, partition := range table.GetPartitions() {
		if partition.GetSize() != 512 && partition.GetStart() != 0 {
			partitions = append(partitions, Partition{
				Partition:       partition,
				Number:          uint32(i + 1),
				AssociatedImage: "",
				LastUsedTime:    time.Now().Unix(),
				DeviceFile:      fmt.Sprintf("/dev/sda%d", uint32(i+1)),
			})
		}
	}

	return partitions
}

// getPartition finds a partition which can be used to store the image. It will either try to find where it was stored prior or use the least recently used partition.
func getPartition(image images.ImageUUID) *Partition {
	chosenPartition := &partitionList[0]

	for i := range partitionList {
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
	chosenPartition.LastUsedTime = time.Now().Unix()
	chosenPartition.AssociatedImage = image
	return chosenPartition
}

// writePartitionJSON writes the partition cache to a file on disk.
func writePartitionJSON(file *os.File) { //nolint
	err := json.NewEncoder(file).Encode(partitionList)

	if err != nil {
		fmt.Printf("Cannot encode JSON to file: %v", err)
	}
}

// printPartitions prints all the partitions in the cache
func printPartitions() {
	for _, partition := range partitionList {
		printPartition(partition)
	}
}

// printPartition prints information about a given partition
func printPartition(partition Partition) {
	log.Debugf("%s %s %d\n",
		partition.DeviceFile, partition.AssociatedImage,
		partition.LastUsedTime)
}
