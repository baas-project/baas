#!/bin/bash
set -e

pushd .

# Update submodule?
cd "$SCRIPT_PATH/linux/"

make mrproper
cp "$SCRIPT_PATH/KERNEL_CONFIG" .config #make defconfig

# Build the kernel
make -j "$(nproc)"

mv "arch/x86/boot/bzImage" "$SCRIPT_PATH/vmlinuz"

popd
