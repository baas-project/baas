FROM golang as build

WORKDIR /build


# Force modules
ENV GO111MODULE=on

# Cache dependencies
COPY go.* ./
RUN go mod download

# Build project
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" ./management_os/programs

# ---------------------

FROM debian:buster
RUN apt-get update
RUN apt-get install -y dhcpcd5
COPY --from=build /build/programs programs
