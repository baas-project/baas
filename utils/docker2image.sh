#!/usr/bin/env sh

# Load our convience functions
. utils/container.sh

cd "$1" || exit
echo -e "*.img\n.*" > .dockerignore
createImage "."
