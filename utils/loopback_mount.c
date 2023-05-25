#include <blkid/blkid.h>
#include <err.h>
#include <fcntl.h>
#include <linux/loop.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <sys/mount.h>
#include <sys/stat.h>
#include <unistd.h>

#define errExit(msg)    do { perror(msg); exit(EXIT_FAILURE);	\
  } while (0)


struct partition_info {
  const char *uuid;
  const char *label;
  const char *type;
};

struct partitions {
  int npartitions;
  struct partition_info *partition_list;
  const char *device;

};

char *get_loop_device() {
  long loopctlfd;
  if ((loopctlfd = open("/dev/loop-control", O_RDWR)) == -1) {
	errExit("open: /dev/loop-control");
  }

  long devnr;
  if ((devnr = ioctl(loopctlfd, LOOP_CTL_GET_FREE)) == -1) {
	errExit("ioctl-LOOP_CTL_GET_FREE");
  }

  char *loopname = malloc(256);
  snprintf(loopname, 256, "/dev/loop%ld", devnr);
  printf("loopname = %s\n", loopname);

  return loopname;
}

void set_loop(long loopfd, long filefd) {
  if (ioctl(loopfd, LOOP_SET_FD, filefd) == -1) {
	errExit("ioctl-LOOP_SET_FD");
  }

  // Set the partition scan flag 
  struct loop_info li;
  if (ioctl(loopfd, LOOP_GET_STATUS, &li) == -1) {
	errExit("ioctl-LOOP_GET_STATUS");
  }
  
  li.lo_flags = LO_FLAGS_PARTSCAN | li.lo_flags;
  
  if (ioctl(loopfd, LOOP_SET_STATUS, &li) == -1) {
  	errExit("ioctl-SET_STATUS");
  }
}

void mount_image_file(long loopfd, const char *filename) {
  long backingfile = -1;
  if ((backingfile = open(filename, O_RDONLY)) == -1) {
	errExit("cannot open image file");
  }

  set_loop(loopfd, backingfile);

  close(backingfile);
}

struct partitions *get_partition_info(long loopfd, char *loopname) {
  blkid_probe pr = blkid_new_probe();
  if (blkid_probe_set_device(pr, loopfd, 0, 0) == -1) {
	errExit("partition_info: failed to set probe");
  }

  blkid_partlist ls = blkid_probe_get_partitions(pr);
  if (ls == NULL) {
	return NULL;
  }

  int npartitions = blkid_partlist_numof_partitions(ls);
  printf("Number of partitions:%d\n", npartitions);

  // A small sleep of 5ms so that the loop devices can be
  // setup. Without this, the code will fail since we try
  // to open the subdevices before they can be created
  usleep(5000);

  struct partitions *partitions = malloc(sizeof(struct partitions));

  partitions->partition_list = calloc(sizeof(struct partition_info), npartitions);
  partitions->npartitions = npartitions;

  for (int i = 0; i < npartitions; i++) {
	struct partition_info partition = partitions->partition_list[i];

	char dev_name[256] = {};
	int n = snprintf(dev_name, 256, "%sp%d", loopname, (i+1));
	dev_name[n] = '\0';

	blkid_probe child_probe = blkid_new_probe_from_filename(dev_name);
	if (child_probe == NULL) {
	  fprintf(stderr, "Invalid child probe device: %s\n", dev_name);
	  continue;
	}

	/*
	  Get the data from the partition and put them into an information struct.
	 */
	blkid_do_probe(child_probe);
	blkid_probe_lookup_value(child_probe, "UUID", (const char**)&partition.uuid, NULL);
	blkid_probe_lookup_value(child_probe, "LABEL", (const char**)&partition.label, NULL);
	blkid_probe_lookup_value(child_probe, "TYPE", (const char**)&partition.type, NULL);

	printf("Name=%s, UUID=%s, LABEL=%s, TYPE=%s\n", dev_name, partition.uuid,
		   partition.label, partition.type);
	blkid_free_probe(child_probe);
  }

  blkid_free_probe(pr);

  return partitions;
}

void free_partitions(struct partitions *partitions) {
  free(partitions->partition_list);
  free((void*)partitions->device);
  free(partitions);
}

struct partitions *get_image_info(const char *image_name) {
  char *loopname = get_loop_device();

  long loopfd = -1;
  if ((loopfd = open(loopname, O_RDWR)) == -1) {
	errExit("open: loopname");
  }

  mount_image_file(loopfd, image_name);
  struct partitions *partitions = get_partition_info(loopfd, loopname);
  if (partitions != NULL) {
	partitions->device = loopname;
  }  else {
	free(loopname);
  }

  close(loopfd);
  return partitions;
}

void mount_partitions(struct partitions *partitions) {
  static const long mount_flags = MS_NOATIME | MS_SILENT | MS_NODEV | MS_NOEXEC | MS_NOSUID;
  static const char* mount_data = "journal_checksum,errors=remount-ro,data=ordered";
  for (int i = 0; i < partitions->npartitions; i++) {
	struct partition_info partition = partitions->partition_list[i];
	char dirname[256];
	snprintf(dirname, 256, "/mnt/partition%d", i);
	mkdir(dirname, 0755);

	char devname[256];
	snprintf(dirname, 256, "%sp%d", partitions->device, i);

	int err = mount(devname, dirname, partition.type,  mount_flags, mount_data);
	if (err != 0) {
	  errExit("mount_partitions: mount");
	}
  }
}

void umount_partitions(struct partitions *partitions) {
  for (int i = 0; i < partitions->npartitions; i++) {
	char dirname[256];
	snprintf(dirname, 256, "/mnt/partition%d", i);

	// Forcefully unmount and delete the left-over folder
	umount2(dirname, MNT_FORCE);
	rmdir(dirname);
  }
}

#ifdef DEBUG
int main(int argc, char **argv) {
  if (argc < 2) {
	fprintf(stderr, "%s requires a filename as argument. Example: %s file.img\n", argv[0], argv[0]);

  }
  struct partitions *partitions = get_image_info(argv[1]);
  if (partitions == NULL) return -1;

//  mount_partitions(partitions);
  free_partitions(partitions);
}
#endif
