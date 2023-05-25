#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>
#include <linux/loop.h>
#include <sys/ioctl.h>
#include "network.c"
#include "cJSON.h"
#include "baas-json.c"
#include <fcntl.h>

#define FILE_PATH "/test.img"

#define errExit(msg)    do { perror(msg); exit(EXIT_FAILURE);	\
  } while (0)

static void sigpoweroff(void);
static void sigreap(void);
static void sigreboot(void);
static void sigchild(int);
static void sigterminate(int);
static void spawn(char *const[]);

static struct {
	int sig;
	void (*handler)(void);
} sigmap[] = {
	{ SIGUSR1, sigpoweroff },
	{ SIGCHLD, sigreap },
	{ SIGALRM, sigreap },
	{ SIGINT, sigreboot },
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

  close(loopctlfd);
  return loopname;
}

static sigset_t set;

static void spawn(char *const argv[]) {
	pid_t pid = fork();
	if (pid < 0) {
		perror("fork"); exit(EXIT_FAILURE);
	} else if (pid == 0) {
		sigprocmask(SIG_UNBLOCK, &set, NULL);
		setsid();
		perror(argv[0]);
		execvp(argv[0], argv);
		_exit(1);
	}

	waitpid(pid, NULL, 0);
}

void get_image(void) {
	pid_t pid = fork();
	if (pid < 0) {
		perror("fork"); exit(EXIT_FAILURE);
	} else if (pid != 0) return;
	for (int i = 1; i <= 31; i++){
		signal(i, sigchild);
	}
	sleep(10);
	puts("STARTS PROCESS");
	sigprocmask(SIG_UNBLOCK, &set, NULL);
	setsid();
	perror("Get network data");
	network_init();
	network_set_uri("192.168.2.33:9090/disk.json");
	char buf[1024];
	network_execute(buf, write_buffer);
	/* network_destroy(); */
	cJSON *json = cJSON_Parse(buf);
	struct baas_setup *bs = parse_baas_setup(json);
	for (int i = 0; i < bs->images_len; i++) {
		struct baas_image_frozen *bif = bs->images[i];
		char uri[256];
		sprintf(uri, "http://192.168.2.33:4848/image/%s/%d", bif->image->uuid, bif->image->version);
		network_set_uri(uri);
		FILE *fptr = fopen(FILE_PATH, "wr");
		network_execute(fptr, write_data);
		fclose(fptr);

		char *loopname = get_loop_device();
		long loopfd = -1;
		if ((loopfd = open(loopname, O_RDWR)) == -1) {
			errExit("open: loopname");
		}

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

		li.lo_flags = LO_FLAGS_PARTSCAN;

		if (ioctl(loopfd, LOOP_SET_STATUS, &li) == -1) {
			errExit("ioctl-LOOP_SET_STATUS");
		}

		close(fd);
		close(loopfd);
	}

	free_baas_setup(bs);
	cJSON_Delete(json);
	_exit(0);
}

int main(void) {
	if (getpid() != 1) {
		return 1;
	}

	chdir("/");
	sigfillset(&set);
	sigprocmask(SIG_BLOCK, &set, NULL);

	/*
	 * Set signal handlers
	 */
	signal(SIGTERM, sigterminate);

	/*
	 * Load the network driver and the loopback device
	 */
	/* char *const modprobe_argv[] = {"modprobe", "loop", NULL}; */
	/* char *const ifup_argv[] = {"ifup", "ens3", NULL}; */
	/* spawn(modprobe_argv); */
	/* spawn(ifup_argv); */

	//get_image();

	chdir("/");
	FILE *fptr = fopen("/baas.log", "w");
	if (fptr == NULL)
		return 1;

	fputs("Hello from the baas init!\n", fptr);
	fclose(fptr);

	extern char *const environ[];
	char *const command[] = {"init", NULL};

	execve("/sbin/init-orig", command, environ);
}

static void sigterminate(int pid) {
	puts("SIGTERM called");
	_exit(0);
}

static void sigchild(int no) {
	pid_t pid;
	int status;
	printf("SIGCHLD called: %d\n", no);
	while ((pid = waitpid(-1, &status, WNOHANG)) > 0) {
		;
	}
}
static void sigpoweroff(void) {

}
static void sigreap(void) {
	while (waitpid(-1, NULL, WNOHANG) > 0)
		;
	alarm(50);
}

static void sigreboot(void) {

}
