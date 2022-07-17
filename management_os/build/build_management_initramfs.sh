#!/usr/bin/env bash
# Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source $SCRIPT_PATH/../../utils/container.sh

# Generate the image directory
generateImage $SCRIPT_PATH

CONTROL_SERVER_IP="$(hostname -I | awk '{print $1}')"

cat > "$SCRIPT_PATH/hosts" << EOF
# Put the ip address of the control server here so the management
# os can start communicating with it once it's booted.
$CONTROL_SERVER_IP        control_server
127.0.0.1                 localhost
EOF

printf "placing /etc/hosts file \033[0;31m(control_server ip set to: $CONTROL_SERVER_IP)\033[0m. To change this edit the CONTROL_SERVER_IP envvar.\n"
cp "$SCRIPT_PATH/hosts" "$SCRIPT_PATH/extract/etc/hosts"

# Place the init script in the extracted folder
echo "placing init script"
#cp "$SCRIPT_PATH/init.sh" "$SCRIPT_PATH/extract/init"

# Copy the kernel 
mv ${SCRIPT_PATH}/extract/boot/vmlinuz-* "$SCRIPT_PATH/../../control_server/static/vmlinuz"
rm ${SCRIPT_PATH}/extract/{vmlinuz.old,initrd.img,initrd.img.old}

# make `init` exec
chmod +x "$SCRIPT_PATH/extract/init" 
pushd .

echo "recompressing initial ramdisk to create initramfs"

# Compress the extracted docker image into a cpio.gz archive for initramfs
cd "$SCRIPT_PATH/extract/"
find . -print0 | cpio --null -o --format=newc | gzip -1 > ../initramfs.cpio.gz

popd

# Rename initramfs
mv "$SCRIPT_PATH/initramfs.cpio.gz" "$SCRIPT_PATH/../../control_server/static/initramfs"

# Cleanup
cleanupImage $SCRIPT_PATH
