mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(shell dirname $(mkfile_path))

# Change this to your own local ip address for testing,
# or to the ip address of the control server when ~~testing~~ running in production.
# This will be put in the hosts file.
export CONTROL_SERVER_IP ?= 192.168.0.142

lint_fix:
	goimports -local baas -w **/*.go
	golangci-lint run --fix

lint:
	goimports -local baas -w **/*.go
	golangci-lint run

management_os: management_initramfs management_kernel

management_initramfs: control_server/static/initramfs
	@$(mkfile_dir)/management_os/build/build_management_initramfs.sh

management_kernel: control_server/static/vmlinuz
	@$(mkfile_dir)/management_os/build/build_management_kernel.sh

control_server_docker:
	@docker-compose -f $(mkfile_dir)/docker-compose.yml up --build

.PHONY: control_server
control_server:
	cd $(mkfile_dir) && sudo env GO111MODULE=on go run ./control_server


destroy_zfs:
	sudo umount /dev/nbd0 || true
	sudo zpool destroy baas || true
	sudo qemu-nbd --disconnect /dev/nbd0 || true

create_zfs:
	mkdir -p control_server/disks
	qemu-img create -f qcow2 baas.qcow2 30G
	sudo modprobe nbd
	sudo qemu-nbd -c /dev/nbd0 baas.qcow2
	# https://openzfs.github.io/openzfs-docs/
	sudo zpool create baas nbd0
	sudo zfs set compression=lz4 baas
	sudo zfs set dedup=on baas
	sudo zfs set recordsize=1M baas

mount_zfs:
	#sudo modprobe nbd
	#sudo qemu-nbd -c /dev/nbd0 baas.qcow2
	sudo zfs set mountpoint=${mkfile_dir}/control_server/disks baas
