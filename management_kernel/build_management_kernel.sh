SCRIPT_PATH=$(dirname "$0")

set -e

docker rm management_kernel || true

docker build -t management_kernel "$SCRIPT_PATH"
CID=$(docker run -d management_kernel /bin/true)

docker export -o "$SCRIPT_PATH/management_kernel_initramfs.tar" ${CID}

mkdir -p "$SCRIPT_PATH/extract"
tar -C "$SCRIPT_PATH/extract" -xf "$SCRIPT_PATH/management_kernel_initramfs.tar"

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

chmod +x "$SCRIPT_PATH/extract/init"
pushd .

cd "$SCRIPT_PATH/extract/"

find . -print0 | cpio --null -ov --format=newc | gzip -9 > ../initramfs.cpio.gz

popd

mkdir -p "$SCRIPT_PATH/kernel"
pushd .
cd "$SCRIPT_PATH/kernel/"

# might not be able to clone if the directory still exists
git clone --depth 1 git://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git || true
cd linux

#cp "$SCRIPT_PATH/KERNEL_CONFIG" .config
make defconfig

make -j $(nproc)

mv "arch/x86/boot/bzImage" "$SCRIPT_PATH/vmlinuz"

popd

rm -rf "$SCRIPT_PATH/extract"
rm -rf "$SCRIPT_PATH/kernel"
rm "$SCRIPT_PATH/management_kernel_initramfs.tar"
mv "$SCRIPT_PATH/initramfs.cpio.gz" "$SCRIPT_PATH/initramfs"
