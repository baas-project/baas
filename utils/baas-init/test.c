#include <stdio.h>
#include <linux/loop.h>
#include <sys/types.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/ioctl.h>
#include <stdlib.h>
#include <stdio.h>
#define FILE_PATH "/test.img"
#define errExit(msg)    do { perror(msg); exit(EXIT_FAILURE);	\
  } while (0)


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

  close(loopctlfd);
  return loopname;
}



int main(void) {
	char *loopname = get_loop_device();
	long loopfd = -1;
	if ((loopfd = open(loopname, O_RDWR)) == -1) {
		errExit("open: loopname");
	}

	puts(loopname);
	long fd;
	if ((fd = open(FILE_PATH, O_RDWR)) == -1) {
		errExit("open: loopname");
	}

	if (ioctl(loopfd, LOOP_SET_FD, fd) == -1) {
		errExit("ioctl-LOOP_SET_FD");
	}

	struct loop_info li;
	if (ioctl(loopfd, LOOP_GET_STATUS, &li) == -1) {
		errExit("ioctl-LOOP_GET_STATUS");
	}

	li.lo_flags = LO_FLAGS_PARTSCAN | LO_FLAGS_DIRECT_IO | ~LO_FLAGS_READ_ONLY;

	if (ioctl(loopfd, LOOP_SET_STATUS, &li) == -1) {
		errExit("ioctl-LOOP_SET_STATUS");
	}

	close(fd);
	close(loopfd);
}
