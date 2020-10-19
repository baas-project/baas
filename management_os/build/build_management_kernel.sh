#!/bin/bash
set -e

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

pushd .

# Update submodule?
cd "$SCRIPT_PATH/linux/"

make mrproper
cp "$SCRIPT_PATH/KERNEL_CONFIG" .config #make defconfig
make olddefconfig

# Build the kernel
make -j "$(nproc)"

mv "arch/x86/boot/bzImage" "$SCRIPT_PATH/../../control_server/static/vmlinuz"

popd
