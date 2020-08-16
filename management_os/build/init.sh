#!/bin/sh

mount -t proc none /proc
mount -t sysfs none /sys

echo Booted succesfully!

exec /bin/sh
