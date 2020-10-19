#!/bin/sh

mount -t proc none /proc
mount -t sysfs none /sys

echo Booted succesfully! $(cut -d' ' -f1 /proc/uptime) seconds

# Run this as a child of init
./programs

# if python ever exists (which it shouldn't)
# run sh replacing init, to debug
exec sh
