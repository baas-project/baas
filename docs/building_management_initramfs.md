# Building the management initramfs

The management initramfs is created based on a docker container. 
With the [dockerfile](../management_os/build/Dockerfile) it's possible to completely
customize the operating system that runs with the [management kernel](building_management_kernel.md)

## prerequisites

1. Docker

# Building

run the [`build_management_kernel.sh`](../management_os/build/build_management_initramfs.sh) file
or run `make management_initramfs` from the project root. 