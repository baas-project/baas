mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(shell dirname $(mkfile_path))

management_os: management_initramfs management_kernel

management_initramfs:
	@$(mkfile_dir)/management_os/build/build_management_initramfs.sh

management_kernel:
	@$(mkfile_dir)/management_os/build/build_management_kernel.sh

control_server_docker: management_initramfs
	@docker-compose -f $(mkfile_dir)/docker-compose.yml up --build

control_server: management_initramfs