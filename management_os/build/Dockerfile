FROM golang as build

WORKDIR /build


# Force modules
ENV GO111MODULE=on

# Cache dependencies
COPY go.* ./
RUN go mod download

# Build project
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" ./management_os/entrypoint

# ---------------------

FROM debian:stable
RUN apt-get update
RUN apt-get install systemd mg parted dhcpcd5 linux-image-amd64 -y
RUN systemctl enable getty@tty1.service
RUN systemctl enable systemd-networkd.service
RUN systemctl enable systemd-networkd-wait-online.service
RUN ln -sr /bin/systemd /init

RUN /bin/echo -e "toor\ntoor" | passwd
RUN rm /etc/localtime && ln -s /usr/share/zoneinfo/Europe/Amsterdam /etc/localtime

COPY --from=build /build/management_os/config/entrypoint.service /etc/systemd/system/entrypoint.service
COPY --from=build /build/management_os/config/20-dhcp.network    /etc/systemd/network/20-dhcp.network
COPY --from=build /build/entrypoint entrypoint

RUN systemctl enable entrypoint
