#!/bin/bash
set -e

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

echo "building docker container"

# Build the docker container
pushd .
cd "$SCRIPT_PATH"/../..
docker build -t management_kernel -f "$SCRIPT_PATH/Dockerfile" .
popd

# Run the container to get a container id
CID=$(docker run -d management_kernel /bin/true)

echo "creating initial ram disk from docker container"

# Export the flat container to a tar
docker export -o "$SCRIPT_PATH/management_kernel_initramfs.tar" "${CID}"

# Extract the tar to ./extract
echo "extracting initial ramdisk"

mkdir -p "$SCRIPT_PATH/extract"
tar -C "$SCRIPT_PATH/extract" -xf "$SCRIPT_PATH/management_kernel_initramfs.tar"

# Place the init script in the extracted folder
echo "placing init script"
cp "$SCRIPT_PATH/init.sh" "$SCRIPT_PATH/extract/init"

CONTROL_SERVER_IP=$(cat $SCRIPT_PATH/hosts | grep -vE "^#|^$" | awk -F " " "{print \$1}")
printf "placing /etc/hosts file \033[0;31m(control_server ip set to: $CONTROL_SERVER_IP)\033[0m\n"
cp "$SCRIPT_PATH/hosts" "$SCRIPT_PATH/extract/etc/hosts"


# make `init` exec
chmod +x "$SCRIPT_PATH/extract/init"
pushd .

echo "recompressing initial ramdisk to create initramfs"

# Compress the extracted docker image into a cpio.gz archive for initramfs
cd "$SCRIPT_PATH/extract/"
find . -print0 | cpio --null -o --format=newc | gzip -9 > ../initramfs.cpio.gz

popd

# Cleanup
rm -rf "$SCRIPT_PATH/extract"
rm "$SCRIPT_PATH/management_kernel_initramfs.tar"

# Rename initramfs
mv "$SCRIPT_PATH/initramfs.cpio.gz" "$SCRIPT_PATH/../../control_server/static/initramfs"
