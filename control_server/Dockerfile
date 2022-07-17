# Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

FROM golang as build

WORKDIR /build

# Force modules
ENV GO111MODULE=on

# Cache dependencies
COPY go.* ./
RUN go mod download
RUN mkdir /static

# Build project
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o control_server_bin ./control_server

# Run stage
FROM gcr.io/distroless/base
WORKDIR /app

COPY --from=build /build/control_server_bin /app/control_server_bin
COPY --from=build /static /static

ENTRYPOINT ["/app/control_server_bin"]
