#!/bin/bash

SCRIPT_PATH=$(dirname "$0")

set -e

# TODO: Is this needed?
docker rm management_kernel || true

# Build the docker container
docker build -t management_kernel "$SCRIPT_PATH"
# Run the container to get a container id
CID=$(docker run -d management_kernel /bin/true)

# Export the flat container to a tar
docker export -o "$SCRIPT_PATH/management_kernel_initramfs.tar" "${CID}"

# Extract the tar to ./extract
mkdir -p "$SCRIPT_PATH/extract"
tar -C "$SCRIPT_PATH/extract" -xf "$SCRIPT_PATH/management_kernel_initramfs.tar"

# Place the init script in the extracted folder
# TODO: Should this be its own file?
cat > "$SCRIPT_PATH/extract/init" <<EOF
#!/bin/sh

mount -t proc none /proc
mount -t sysfs none /sys

cat <<!

Boot took $(cut -d' ' -f1 /proc/uptime) seconds

Welcome to your docker image based linux.
!
exec /bin/sh

EOF

# make `init` exec
chmod +x "$SCRIPT_PATH/extract/init"
pushd .

# Compress the extracted docker image into a cpio.gz archive for initramfs
cd "$SCRIPT_PATH/extract/"
find . -print0 | cpio --null -ov --format=newc | gzip -9 > ../initramfs.cpio.gz

popd

mkdir -p "$SCRIPT_PATH/kernel"
pushd .
cd "$SCRIPT_PATH/kernel/"

# Update submodule?
cd linux

make mrproper
cp "$SCRIPT_PATH/KERNEL_CONFIG" .config
#make defconfig

# Build the kernel
make -j "$(nproc)"

mv "arch/x86/boot/bzImage" "$SCRIPT_PATH/vmlinuz"

popd

# Cleanup
rm -rf "$SCRIPT_PATH/extract"
rm -rf "$SCRIPT_PATH/kernel"
rm "$SCRIPT_PATH/management_kernel_initramfs.tar"

# Rename initramfs
mv "$SCRIPT_PATH/initramfs.cpio.gz" "$SCRIPT_PATH/initramfs"
