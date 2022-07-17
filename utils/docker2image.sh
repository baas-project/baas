#!/usr/bin/env sh
# Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.


# Load our convience functions
. utils/container.sh

cd "$1" || exit
echo -e "*.img\n.*" > .dockerignore
createImage "."
