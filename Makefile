mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(shell dirname $(mkfile_path))

# Change this to your own local ip address for testing,
# or to the ip address of the control server when ~~testing~~ running in production.
# This will be put in the hosts file.
export CONTROL_SERVER_IP ?= 192.168.2.75

lint_fix:
	goimports -local baas -w **/*.go
	golangci-lint run --fix

lint:
	goimports -local baas -w **/*.go
	golangci-lint run

management_os: management_initramfs management_kernel

management_initramfs:
	@$(mkfile_dir)/management_os/build/build_management_initramfs.sh

management_kernel:
	@$(mkfile_dir)/management_os/build/build_management_kernel.sh

management_initramfs: control_server/static/initramfs

management_kernel: control_server/static/vmlinuz

control_server_docker:
	@docker-compose -f $(mkfile_dir)/docker-compose.yml up --build

.PHONY: control_server
control_server:
	cd $(mkfile_dir) && sudo env GO111MODULE=on go run ./control_server
