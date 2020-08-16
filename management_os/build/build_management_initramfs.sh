#!/bin/bash
set -ex

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Build the docker container
pushd .
cd "$SCRIPT_PATH"/..
docker build -t management_kernel "$SCRIPT_PATH"
popd

# Run the container to get a container id
CID=$(docker run -d management_kernel /bin/true)

# Export the flat container to a tar
docker export -o "$SCRIPT_PATH/management_kernel_initramfs.tar" "${CID}"

# Extract the tar to ./extract
mkdir -p "$SCRIPT_PATH/extract"
tar -C "$SCRIPT_PATH/extract" -xf "$SCRIPT_PATH/management_kernel_initramfs.tar"

# Place the init script in the extracted folder
cp "$SCRIPT_PATH/init.sh" "$SCRIPT_PATH/extract/init"

# make `init` exec
chmod +x "$SCRIPT_PATH/extract/init"
pushd .

# Compress the extracted docker image into a cpio.gz archive for initramfs
cd "$SCRIPT_PATH/extract/"
find . -print0 | cpio --null -ov --format=newc | gzip -9 > ../initramfs.cpio.gz

popd

# Cleanup
rm -rf "$SCRIPT_PATH/extract"
rm "$SCRIPT_PATH/management_kernel_initramfs.tar"

# Rename initramfs
mv "$SCRIPT_PATH/initramfs.cpio.gz" "$SCRIPT_PATH/../../static/initramfs"
