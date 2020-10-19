mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(shell dirname $(mkfile_path))

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