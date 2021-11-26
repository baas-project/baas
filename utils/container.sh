#!/usr/bin/env bash

function generateImage {
	echo "building docker container"

	# Build the docker container
	docker build -t container -f "$1/Dockerfile" .

	# Run the container to get a container id
	CID=$(docker run -d container /bin/true)

	echo "creating initial ram disk from docker container"

	# Export the flat container to a tar
	docker export -o "$1/container.tar" "${CID}"

	# Extract the tar to ./extract
	echo "extracting initial ramdisk"

	mkdir -p "$1/extract"
	tar -C "$1/extract" -xf "$1/container.tar"
}

function cleanupImage {
    echo "clean the files"
    
    rm -rf "$1/extract"
    rm -f "$1/container.tar"
}

function createDisk { 
    # Create an image made out of 2048 blocks of 4 megabytes
    fallocate -l 8GiB "$1/image.img"

    losetup "$2" "$1/image.img"
    # Turn the image into a disk with one bootable 10 Gigabyte image 
    parted "$2" mklabel gpt
    yes 'I' | parted "$2" mkpart image fat32 0 512M
    yes 'I' | parted "$2" mkpart image ext4 512M 8G
    parted "$2" set 1 boot on
    parted "$2" set 1 esp on
    

    yes | mkfs.fat -F 32 "${2}"p1
    yes | mkfs.ext4 "${2}"p2
}

function populateDisk { 
    mkdir mnt
    mount "${2}p2" mnt/
    mkdir mnt/boot
    mount "${2}p1" mnt/boot
    
    cp -r "${1}"/extract/* mnt/
    mount --types proc /proc mnt/proc
    mount --rbind /sys mnt/sys
    mount --make-rslave mnt/sys
    mount --rbind /dev mnt/dev
    mount --make-rslave mnt/dev
    mount --bind /run mnt/run
    mount --make-rslave mnt/run
    
    chroot mnt sh -c 'bootctl install'
    /usr/bin/echo -e "default arch.conf
timeout 4
console-mode max
editor no    
" > mnt/boot/loader/loader.conf

    echo -e "title   Arch Linux
linux   /vmlinuz-linux
initrd  /initramfs-linux.img
options root=/dev/sda2 rw    
" > mnt/boot/loader/entries/arch.conf

    umount -l mnt/{sys,dev,run,boot} mnt
    
    rm -Rf mnt/boot
    rm -Rf mnt
}

function createImage {
    LOOPDEVICE=$(losetup -f)
    
    generateImage "$1" 
    createDisk "$1"  "$LOOPDEVICE"
    populateDisk "$1" "$LOOPDEVICE"
    cleanupImage "$1"
    
    qemu-system-x86_64 -drive file=image.img,index=0,media=disk,format=raw -bios /usr/share/edk2/ovmf/OVMF_CODE.fd -m 2G
}
