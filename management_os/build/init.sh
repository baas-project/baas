#!/bin/bash

mount -t proc none /proc
mount -t sysfs none /sys

echo Requesting IP address

dhcpcd

sleep 10

echo Booted succesfully! $(cut -d' ' -f1 /proc/uptime) seconds

# Run this as a child of init
./programs

# if go ever exists (which it shouldn't)
# run sh replacing init, to debug
exec bash
